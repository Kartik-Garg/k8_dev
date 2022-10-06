package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "/home/kartik/.kube/config", "location for kubeconfig file")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("Error: %s, getting config", err.Error())
		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Printf("Error:%s", err.Error())
		}
	}
	//this gets us dynamic client
	dynclient, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error: %s, in getting dynamic client", err.Error())
	}

	unObject, err := dynclient.Resource(schema.GroupVersionResource{
		Group:    "kartikgarg.dev",
		Version:  "v1alpha1",
		Resource: "etcds",
	}).Namespace("default").Get(context.Background(), "etcd-0", metav1.GetOptions{})
	if err != nil {
		fmt.Printf("We got error:%s, while getting custom resources", err.Error())
	}
	//fmt.Printf("Length of the list of items &%d\n", len(resources.Items))

	//creating new dynamic infromer
	infFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynclient, 10*time.Minute)
	//creating the new controller
	c := newController(dynclient, infFactory)
	infFactory.Start(make(<-chan struct{}))
	c.run(make(<-chan struct{}))

	fmt.Printf("Get the object: %s\n", unObject.GetName())
}
