package main

import (
	"flag"
	"fmt"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfigfile", "/home/kartik/.kube/config", "This is where config file is located")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("We have error:%s, config file", err.Error())

		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Printf("Error %s, in cluster config", err.Error())
		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error:%s, creating client set", err.Error())
	}

	informers := informers.NewSharedInformerFactory(clientset, 10*time.Minute)
	fmt.Print(informers)
}
