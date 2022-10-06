package main

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

type controller struct {
	client dynamic.Interface
	//need informer for the run function
	informer cache.SharedIndexInformer
}

func newController(client dynamic.Interface, dynInfromer dynamicinformer.DynamicSharedInformerFactory) *controller {

	inf := dynInfromer.ForResource(schema.GroupVersionResource{
		Group:    "kartikgarg.dev",
		Version:  "v1alpha1",
		Resource: "etcds",
	}).Informer()

	inf.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				fmt.Println("New etcd resource was created")
			},
			DeleteFunc: func(obj interface{}) {
				fmt.Print("New Etcd resource was deleted\n")
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				fmt.Print("etcd resource was updated")
			},
		},
	)

	return &controller{
		client:   client,
		informer: inf,
	}
}

func (c *controller) run(ch <-chan struct{}) {
	fmt.Println("Starting the controller now")
	if !cache.WaitForCacheSync(ch, c.informer.HasSynced) {
		//if not synced
		fmt.Println("waiting for cache to finish up the sync")
	}

	go wait.Until(c.worker, 1*time.Second, ch)

	<-ch
}

func (c *controller) worker() {
	for c.processItem() {

	}
}

func (c *controller) processItem() bool {
	return true
}
