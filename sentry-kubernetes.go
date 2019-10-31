package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/getsentry/sentry-go"
	api "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // external cluster config

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var debugFlag = flag.Bool("debug", false, "Enable debug logging --debug true")

func main() {
	flag.Parse()
	// config, err := rest.InClusterConfig()

	var kubeconfig *string
	if home := os.Getenv("HOME"); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	dsn := os.Getenv("DSN")

	if dsn == "" {
		fmt.Println("Missing DSN ENV token")
		os.Exit(1)
	}

	namespace := os.Getenv("namespace")

	if namespace == "" {
		namespace = api.NamespaceAll
	}
	env := os.Getenv("ENV")

	err = sentry.Init(sentry.ClientOptions{
		Dsn:         dsn,
		Environment: env,
	})
	if err != nil {
		fmt.Println("unable to connect to sentry")
		os.Exit(1)
	}

	fmt.Println("Starting go-sentry-kubernetes")
	debug(fmt.Sprintf("Using ENV: %s\n", env))

	//clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	watchlist := cache.NewListWatchFromClient(
		clientset.Core().RESTClient(),
		"pods",
		namespace,
		fields.Everything(),
	)

	_, controller := cache.NewInformer(
		watchlist,
		&api.Pod{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: handleEvent,
		},
	)
	queue := make(chan struct{})
	go controller.Run(queue)
	select {}
}

func debug(msg string) {
	if *debugFlag {
		fmt.Println(msg)
	}
}

func handleEvent(_, obj interface{}) {
	pod := obj.(*api.Pod)
	statuses := pod.Status.ContainerStatuses
	b, _ := json.MarshalIndent(obj, "", "  ")
	fmt.Println(string(b))

	for _, status := range statuses {
		if status.LastTerminationState != (api.ContainerState{}) {
			exitCode := status.LastTerminationState.Terminated.ExitCode
			codeStr := fmt.Sprintf("%d", exitCode)
			containerMessage := ""
			containerReason := ""

			if status.LastTerminationState.Terminated != nil && exitCode != 0 {
				containerMessage = status.LastTerminationState.Terminated.Message
			}
			if status.LastTerminationState.Terminated.Reason == "Completed" {
				// break
			}
			if status.LastTerminationState.Terminated.Reason != "" {
				if containerMessage != "" {
					containerMessage += " - "
				}
				containerMessage += status.LastTerminationState.Terminated.Reason
			}
			if containerMessage == "Error" {
				containerReason = containerMessage
				containerMessage = fmt.Sprintf("Pod: %s %s %s", pod.ObjectMeta.Name, "exited with code: ", codeStr)
			}
			if containerMessage == "OOMKilled" {
				containerReason = containerMessage
				containerMessage = fmt.Sprintf("Pod: %s %s", pod.ObjectMeta.Name, "OOMKilled")
			}
			evt := sentry.NewEvent()
			evt.Message = containerMessage
			evt.Release = status.Image
			evt.Platform = "kubernetes"
			evt.Level = sentry.LevelError

			evt.Extra["name"] = pod.Name
			evt.Extra["reason"] = containerReason
			evt.Extra["nodeName"] = pod.Spec.NodeName
			evt.Extra["exitCode"] = codeStr
			evt.Extra["container"] = pod.Spec.Containers[0].Name
			evt.Extra["namespace"] = pod.ObjectMeta.Namespace
			evt.Extra["restartCount"] = status.RestartCount
			sentry.CaptureEvent(evt)
		}
	}
}
