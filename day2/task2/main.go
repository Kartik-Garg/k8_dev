package main

import (
	"flag"
	"fmt"
	"time"

	controller "github.com/kartik-garg/cc/controllers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "/home/kartik/.kube/config", "this is the location for your k8 config file")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("Got an error:%s, while creating the kubeconfig file", err.Error())

		//need to get the config file from inside the cluster, if this application is deployed on a cluster
		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Printf("Got an error:%s, while getting the config file from inside the cluster", err.Error())
		}
	}

	//creating a clientset through config file
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error %s, while creating clientset", err.Error())
	}

	ch := make(chan struct{})
	//need to create informet to keep tabs on the dep resources and then the controller can take actions on it

	//informers := informers.NewSharedInformerFactory(clientset, 10*time.Minute)

	labelOptions := informers.WithTweakListOptions(func(opts *metav1.ListOptions) {
		opts.LabelSelector = "app=k8s-dev"
	})

	informers := informers.NewSharedInformerFactoryWithOptions(clientset, 10*time.Minute, informers.WithNamespace(""), labelOptions)
	//initialize the controller
	c := controller.NewController(clientset, informers.Apps().V1().Deployments())
	//start the informer as well
	informers.Start(ch)
	//run the controller now
	c.Run(ch)

}
