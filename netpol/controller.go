package main

import (
	"fmt"

	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	appsInformer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	appslister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type controller struct {
	//need client set
	clientset kubernetes.Interface
	//need to list all the resources on which we have applied the informer
	podLister appslister.PodLister
	//check if the informer has been synced
	podSync cache.InformerSynced
}

func NewController(clientset kubernetes.Interface, podInformer appsInformer.PodInformer) *controller {
	c := &controller{
		clientset: clientset,
		podLister: podInformer.Lister(),
		podSync:   podInformer.Informer().HasSynced,
	}

	//functions to do when edited the pod
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			fmt.Println("can not change label for this pod as it is being used in the netpol")
		},
	})

	return c
}

func (c *controller) Run(ch <-chan struct{}) {
	fmt.Println("Starting controller")
	//have to check if the cache has been synced for the informer
	if !cache.WaitForCacheSync(ch, c.podSync) {
		//if the cache has not been synced
		fmt.Println("waiting for cache to get syncd")
	}

	//write a function which will keep on working till the channel is closed
	go wait.Until(c.worker, 1*time.Second, ch)

	<-ch
}

func (c *controller) worker() {

}
