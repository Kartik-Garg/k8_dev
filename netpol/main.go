package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "/home/kartik/.kube/config", "This is where the kube config file resides")

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("Error building config from kubeconfig: %s", err.Error())

		//building from internal cluser
		config, err = rest.InClusterConfig()
	}

	//lets build the kubeclient set now
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error while creating the kube client set: %s", err.Error())
	}
	fmt.Println(clientSet)

	podsList, err := clientSet.CoreV1().Pods("net").List(context.Background(), metav1.ListOptions{})
	for _, pods := range podsList.Items {
		fmt.Println(pods.Name)
	}

	var mapOfLabelsInNetPol = make(map[string]string)
	//list all the network policies as well present in the namespace
	netpolList, err := clientSet.NetworkingV1().NetworkPolicies("net").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error while getting the netpolList: %s", err.Error())
	}
	for _, netpol := range netpolList.Items {
		fmt.Println(netpol.Name)
		mapOfLabelsInNetPol = netpol.Spec.PodSelector.MatchLabels

	}

	fmt.Println(mapOfLabelsInNetPol)

	var k string
	var v []string
	for key, value := range mapOfLabelsInNetPol {
		k = key
		v = append(v, value)
	}

	req, err := labels.NewRequirement(k, selection.Equals, v)
	if err != nil {
		fmt.Printf("Error creating new requirements: %s", err.Error())
	}
	selector := labels.NewSelector()
	selector = selector.Add(*req)
	informer := informers.NewSharedInformerFactoryWithOptions(clientSet, 10*time.Minute, informers.WithTweakListOptions(func(lo *metav1.ListOptions) {
		lo.LabelSelector = selector.String()
	}))
	fmt.Println(informer)

	ch := make(chan struct{})

	c := NewController(clientSet, informer.Core().V1().Pods())

	informer.Start(ch)

	c.Run(ch)

}
