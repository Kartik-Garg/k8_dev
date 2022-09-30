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
	//step1: we need kubeconfig file to get access to the api server to hit the requests
	kubeconfig := flag.String("kubeconfig", "/home/kartik/.kube/config", "location to your kube config file.")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("got error while getting kubeconfig file %s/n", err.Error())
	}

	//step2: get client set from config file, we can use named client set since ingress is part of it
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("got error while creating the clientset %s\n", err.Error())
	}

	// we can check which api version our resource belongs to, while creating the ingress we used networking group
	//machinery is used for listoptions

	ingresses, err := clientset.NetworkingV1().Ingresses("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("got an error: %s, while getting the ingresses", err.Error())
	}

	for _, ingress := range ingresses.Items {
		fmt.Printf("Ingresses present are : %s\n", ingress.Name)
	}

	//getting it in the same format as k get ingresses
	// for _, ingress := range ingresses.Items {
	// 	fmt.Print(ingress.Name, ingress.Spec.TLS)
	// }

	result, err := clientset.NetworkingV1().Ingresses("default").Get(context.Background(), "example-ingress", metav1.GetOptions{})
	fmt.Println(result.Name, result.Spec.IngressClassName, result.Spec.TLS, result.Status.LoadBalancer.Ingress, result.Spec.TLS)

	//NOT ABLE TO GET PORT AND AGE OF THE INGRESS

}
