package main

import (
	"flag"
	"fmt"
	"os"
	// "k8s.io/apimachinery/pkg/api/resource"
	api "k8s.io/api/core/v1"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // external cluster config
	// "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"path/filepath"
	// "k8s.io/client-go/rest"
	"time"

	// "k8s.io/client-go/rest"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

func main() {
	// inside the cluster:
	// config, err := rest.InClusterConfig()
	// if err != nil {
	// 	panic(err.Error())
	// }

	var kubeconfig *string
	if home := os.Getenv("HOME"); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	watchlist := cache.NewListWatchFromClient(
		clientset.Core().RESTClient(),
		"pods",
		"product",
		fields.Everything(),
	)

	_, controller := cache.NewInformer(
		watchlist,
		&api.Pod{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				t := obj.(*api.Pod)
				if t.Name == "product-web-1862d89b48d4667af6121ad21947591dfd217480-85c8cw6klw" {
					fmt.Printf("add: %s %s \n", t.Name, t.Status.Phase)
				}
			},
			DeleteFunc: func(obj interface{}) {
				fmt.Printf("delete: s \n")
				t := obj.(*api.Pod)
				if t.Name == "product-web-1862d89b48d4667af6121ad21947591dfd217480-85c8cw6klw" {
					fmt.Printf("%s \n", t.Name)
					fmt.Printf("%s \n", t.Status.Phase)
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				fmt.Printf("old: s, new: s \n")
				t := oldObj.(*api.Pod)
				if t.Name == "product-web-1862d89b48d4667af6121ad21947591dfd217480-85c8cw6klw" {
					fmt.Printf("%s \n", t.Name)
					fmt.Printf("%s \n", t.Status.Phase)

					t2 := newObj.(*api.Pod)
					fmt.Printf("%s \n", t.Name)
					fmt.Printf("%s %s\n", t.Status, t2.Status.Phase)
				}
			},
		},
	)
	stop := make(chan struct{})
	go controller.Run(stop)
	select {}
}
