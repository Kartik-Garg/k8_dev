package controller

import (
	"fmt"

	appinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	applisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type controller struct {
	clientset     kubernetes.Interface
	depLister     applisters.DeploymentLister
	depCacheSyncd cache.InformerSynced
	queue         workqueue.RateLimitingInterface
}

func newController(clientset kubernetes.Interface, depInformer appinformers.DeploymentInformer) *controller {
	c := &controller{
		clientset:     clientset,
		depLister:     depInformer.Lister(),
		depCacheSyncd: depInformer.Informer().HasSynced,
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ekspose"),
	}

	depInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    handleAdd,
			DeleteFunc: handleDel,
		},
	)
	return c
}

func (c *controller) run(ch <-chan struct{}) {
	fmt.Println("Starting the controller")
	if !cache.WaitForCacheSync(ch, c.depCacheSyncd) {
		fmt.Println("Cahce is yet to sync")
	}
}

func handleAdd(obj interface{}) {

}

func handleDel(obj interface{}) {

}
