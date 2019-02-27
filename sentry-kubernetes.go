package main

import (
	"fmt"
	"github.com/getsentry/raven-go"
	api "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // external cluster config
	"k8s.io/client-go/rest"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

var namespace = "product"

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	dsn := os.Getenv("DSN")

	if dsn == "" {
		fmt.Println("Missing DSN ENV token")
		os.Exit(1)
	}
	client, err := raven.New(dsn)
	if err != nil {
		panic("unable to connect to sentry")
	}
	client.SetEnvironment(os.Getenv("ENV"))

	fmt.Println("Starting go-sentry-kubernetes")

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	watchlist := cache.NewListWatchFromClient(
		clientset.Core().RESTClient(),
		"pods",
		api.NamespaceAll,
		fields.Everything(),
	)

	_, controller := cache.NewInformer(
		watchlist,
		&api.Pod{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: func(oldObj, newObj interface{}) {
				t := newObj.(*api.Pod)
				statuses := t.Status.ContainerStatuses
				errorMessage := ""
				var count int32
				for _, status := range statuses {
					if status.LastTerminationState != (api.ContainerState{}) {
						reason := status.LastTerminationState.Terminated.Reason
						containerReason := ""
						if status.State.Terminated != (&api.ContainerStateTerminated{}) && status.State.Terminated != nil {
							containerReason = status.State.Terminated.Reason
						}
						if t.Status.Reason != "" && t.Status.Reason != reason {
							reason = fmt.Sprintf("%s %s", reason, t.Status.Reason)
						}
						errorMessage = fmt.Sprintf("%s %s %s ", errorMessage, reason, containerReason)
						if status.RestartCount != 0 {
							count = status.RestartCount
						}
					}
				}
				if errorMessage != "" {
					message := fmt.Sprintf("%s - %s", errorMessage, t.Name)
					notifySentry(client, errorMessage, message, count)
				}
			},
		},
	)
	queue := make(chan struct{})
	go controller.Run(queue)
	select {}
}

func notifySentry(client *raven.Client, title string, message string, count int32) {
	messages := map[string]string{
		"message":      message,
		"restartCount": fmt.Sprintf("%d", count),
	}
	fmt.Printf("reporting: %s", title)
	raven.CaptureMessage(title, messages)
}
