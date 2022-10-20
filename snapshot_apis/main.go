package main

import (
	"context"
	"flag"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	//for controller first we need to get the kube config, to communicate
	kubeconfig := flag.String("kubeconfig", "/home/kartik/.kube/config", "This is the location for the kube config file")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("We got an error %s, while getting config file from the location", err.Error())

		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Printf("got error while getting in cluster config, %s", err.Error())
		}
	}

	//creating  dynamic clientset now
	dynclient, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error while creating dynamic client %s", err.Error())
	}

	unObject, err := dynclient.Resource(schema.GroupVersionResource{
		Group:    "snapshot.storage.k8s.io",
		Version:  "v1",
		Resource: "volumesnapshots",
	}).Namespace("snapshot").List(context.Background(), metav1.ListOptions{})

	if err != nil {
		fmt.Printf("Getting error while listing the snapshot apis %s", err.Error())
	}
	fmt.Println(unObject)

	//now we can communicate with the cluster, we need to build a controller which can take snapshot and restore them
	//so controller needs to have access to the clientset and snapshots and pvcs as well
	//since the snapshots are CRDS, do we need a custom controller and custom config? - yes
	//needed -> we need to create a snapshot and then restore snapshot (pvc will be used here where resource will be a snapshot)
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
	getCSIDriver()
}
