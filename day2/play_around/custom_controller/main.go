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
	kubeconfig := flag.String("kubeconfig", "/home/kartik/.kube/config", "lcoation for the kube config file")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("got the error %s, while building config", err.Error())

		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Printf("error %s, getting cluster config", err.Error())
		}
	}

	//creating the client set
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("error %s, while creating the client set", err.Error())
	}

	ch := make(chan struct{})

	informers := informers.NewSharedInformerFactory(clientset, 10*time.Minute)
	c := newController(clientset, informers.Apps().V1().Deployments())
	informers.Start(ch)
	c.run(ch)
	fmt.Print(informers)
}
