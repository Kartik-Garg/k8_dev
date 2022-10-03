package main

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"

	"k8s.io/client-go/kubernetes"
	applisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type controller struct {
	//this is the controller which is going to have all the methods that we need to use
	//need client set to interact with k8 cluster
	clientset kubernetes.Interface
	//to list the deployment resources on which the registered functions would be running
	deplister applisters.DeploymentLister
	//check if cache has been synced
	depCacheSyncd cache.InformerSynced
	//need a queue which stores the objects if the registered functions have been called
	queue workqueue.RateLimitingInterface
}

// function which calls controller so we can call it from main and it can return us controller
func newController(clientset kubernetes.Interface, depInformer appsinformers.DeploymentInformer) *controller {
	c := &controller{
		clientset:     clientset,
		deplister:     depInformer.Lister(),
		depCacheSyncd: depInformer.Informer().HasSynced,
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ekspose"),
	}

	depInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAdd,
			DeleteFunc: c.handleDel,
		},
	)

	return c
}

func (c *controller) run(ch <-chan struct{}) {
	fmt.Println("Starting controller")
	if !cache.WaitForCacheSync(ch, c.depCacheSyncd) {
		//we have to wait for the internal cache to initialise and/or sync
		fmt.Print("Error waiting for cache to be synced")
	}
	//calls a specific function, every period till this channel (funciton) is closed
	go wait.Until(c.worker, 1*time.Second, ch)
	//making go routine blocking
	<-ch
}

func (c *controller) worker() {
	//do logic stuff once add and delete are called
	for c.processItem() {

	}

}

func (c *controller) processItem() bool {
	//actual implementation of logic
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}

	//if everything goes fine we also have to remove item from the queue
	defer c.queue.Forget(item)
	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		fmt.Printf("Getting key from cache:%s\n", err.Error())
	}

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		fmt.Printf("Got Error while splitting namespace:%s\n", err.Error())
	}

	err = c.syncDeployment(ns, name)
	if err != nil {
		// re-try
		fmt.Printf("Error while syncing ns: %s\n", err.Error())
		return false
	}
	return true
}

func (c *controller) syncDeployment(ns, name string) error {

	//get deployment from lister
	dep, err := c.deplister.Deployments(ns).Get(name)
	if err != nil {
		fmt.Printf("Got error while getting dpeloyments dfrom liuster: %s", err.Error())
	}

	//create service
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dep.Name,
			Namespace: ns,
		},
		Spec: corev1.ServiceSpec{
			Selector: deplLabels(*dep),
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: 80,
				},
			},
		},
	}
	_, err = c.clientset.CoreV1().Services(ns).Create(context.Background(), &svc, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("creating service, error: %s\n", err.Error())
	}
	//create ingress for the deployment that was created
	return nil

}

func deplLabels(dep appsv1.Deployment) map[string]string {
	return dep.Spec.Template.Labels
}

func (c *controller) handleAdd(obj interface{}) {
	fmt.Println("Add was called")
	c.queue.Add(obj)
}

func (c *controller) handleDel(obj interface{}) {
	fmt.Println("Delete was called")
	c.queue.Add(obj)
}
