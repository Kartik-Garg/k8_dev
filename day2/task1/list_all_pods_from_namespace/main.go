package main

import (
	"context"
	"flag"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	//step1: get the kubeconfig file
	kubeconfig := flag.String("kubeconfig", "/home/kartik/.kube/config", "this is the location for the kubeconfig file")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("Got an error: %s, while creating the config file", err.Error())
	}
	//creating the client set
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("got an error: %s, while creating the client set", err.Error())
	}

	podslist, err := clientset.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Getting an error: %s , while getting the list of pods from the default namespace", err.Error())
	}

	fmt.Print("List of all the pods present in the default namespace are:\n")
	for _, pods := range podslist.Items {
		fmt.Println(pods.Name)
	}
}
