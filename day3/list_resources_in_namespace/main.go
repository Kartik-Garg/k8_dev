package main

import (
	"context"
	"flag"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "/home/kartik/.kube/config", "this is the location for the config file")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("Got an error %s, while creating the kubeconfig file", err.Error())

		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Printf("got an error %s, while building in cluster config", err.Error())
		}
	}
	//create client set to communitcate with the k8s cluster
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Got an error %s while getting the clientset", err.Error())
	}

	var input_resource string
	input_namespace := ""

	// resourcelist, err := clientset.AppsV1().Deployments("").List(context.Background(), metav1.ListOptions{})

	// for _, resource := range resourcelist.Items {
	// 	fmt.Println(resource.Name)
	// }

	fmt.Println("Enter the resouce name")
	fmt.Scanln(&input_resource)
	if input_resource == "deployments" {
		deplist, err := clientset.AppsV1().Deployments(input_namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Error: %s, while getting the list of deployments", err.Error())
		}
		for _, deps := range deplist.Items {
			fmt.Print(deps.Name)
		}
	}

	if input_resource == "services" {
		servicelist, err := clientset.CoreV1().Services(input_namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Error: %s, while getting the list of services", err.Error())
		}
		for _, service := range servicelist.Items {
			fmt.Println(service.Name)
		}
	}

	if input_namespace == "pods" {
		podlist, err := clientset.CoreV1().Pods(input_namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Error: %s, while getting the list of podlists", err.Error())
		}
		for _, pods := range podlist.Items {
			fmt.Println(pods.Name)
		}
	}

}
