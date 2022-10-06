package controller

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	appslister "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
)

type controller struct {
	//what will the controller consist, client set to communicate and make changes
	clientset kubernetes.Interface
	//lister is the component of informer which is used to list all the resources
	dplister appslister.DeploymentLister
	//also need to check if the cache from where the informer is getting is syncd or not
	depCacheSyncd cache.InformerSynced
}

// function which can be called to initialise and get the controller
func NewController(clientset kubernetes.Interface, dpinformer appsinformers.DeploymentInformer) *controller {
	//need to initialize it here
	c := &controller{
		clientset:     clientset,
		dplister:      dpinformer.Lister(),
		depCacheSyncd: dpinformer.Informer().HasSynced,
	}

	//since informer keeps listening to the resources, we need to register the register the required
	//functions, so it can inform us when they take place
	dpinformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    handleAdd,
			DeleteFunc: handleDel,
			UpdateFunc: handleUpd,
		},
	)

	return c
}

func (c *controller) Run(ch <-chan struct{}) {
	fmt.Println("Starting controller")
	//have to check if the cache has been synced for the informer
	if !cache.WaitForCacheSync(ch, c.depCacheSyncd) {
		//if the cache has not been synced
		fmt.Println("waiting for cache to get syncd")
	}

	//write a function which will keep on working till the channel is closed
	go wait.Until(c.worker, 1*time.Second, ch)

	<-ch
}

func (c *controller) worker() {

}

// so whenever add and delete operations take place the events registered with the informers trigger those functions
func handleAdd(obj interface{}) {
	//map to check
	// labelToCheck := map[string]string{"app": "k8s-dev"}
	// deps := obj.(*appsv1.Deployment)
	// if reflect.DeepEqual(labelToCheck, deps.Labels) {

	// 	//fmt.Println(deps)
	// 	fmt.Println("Deployment reource has been added")
	// }
	fmt.Println("Deployment reource has been added")
}

func handleDel(obj interface{}) {
	// labelToCheck := map[string]string{"app": "k8s-dev"}
	// deps := obj.(*appsv1.Deployment)
	// if reflect.DeepEqual(labelToCheck, deps.Labels) {
	// 	fmt.Println("Deployment resource has been removed")
	// }
	fmt.Println("Deployment resource has been removed")
}

func handleUpd(oldObj interface{}, newObj interface{}) {
	// labelToCheck := map[string]string{"app": "k8s-dev"}
	// deps := newObj.(*appsv1.Deployment)
	// if reflect.DeepEqual(labelToCheck, deps.Labels) {
	// 	fmt.Println("Deployment resource has been removed")
	// }
	fmt.Println("Deployment resource has been removed")
}
