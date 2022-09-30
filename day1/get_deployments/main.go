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
	//first step is to get the config file
	kubeconfig := flag.String("kubeconfig", "/home/kartik/.kube/config", "this is the locatio for the kubeconfig file")

	//create the config for the go now
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("Error: %s occured in creating a config for go", err.Error())
	}

	//create client set so we can hit up the api server
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error: %s, occured while creating the clientset", err.Error())
	}

	//use machinery for listoptions
	deployments, err := clientset.AppsV1().Deployments("default").List(context.Background(), metav1.ListOptions{})

	//listing deployment name and the tag
	for _, deployment := range deployments.Items {
		fmt.Printf("deployment name is:%s\ndeployment tag is:%s\n", deployment.Name, deployment.Labels)
	}

	//updating the deployment named deployment tag from app:deployment to app:example
	//we can use get here just as we use with kubernetes to get a particular deployment
	result, err := clientset.AppsV1().Deployments("default").Get(context.Background(), "deployment", metav1.GetOptions{})
	if err != nil {
		fmt.Printf("got error: %s while getting the single deployment", err.Error())
	}

	result.Labels = map[string]string{"updatedlabel": "example"}

	fmt.Printf("updated label is: %s\n", result.Labels)

	//fmt.Println(deployments)
}
