package controller

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	appslister "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type controller struct {
	//what will the controller consist, client set to communicate and make changes
	clientset kubernetes.Interface
	//lister is the component of informer which is used to list all the resources
	dplister appslister.DeploymentLister
	//also need to check if the cache from where the informer is getting is syncd or not
	depCacheSyncd cache.InformerSynced
	//need to have queue as well
	queue workqueue.RateLimitingInterface
}

// function which can be called to initialise and get the controller
func NewController(clientset kubernetes.Interface, dpinformer appsinformers.DeploymentInformer) *controller {
	//need to initialize it here
	c := &controller{
		clientset:     clientset,
		dplister:      dpinformer.Lister(),
		depCacheSyncd: dpinformer.Informer().HasSynced,
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ekspose"),
	}

	//since informer keeps listening to the resources, we need to register the register the required
	//functions, so it can inform us when they take place
	dpinformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAdd,
			DeleteFunc: c.handleDel,
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
	for c.processItem() {

	}
}

func (c *controller) processItem() bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}

	defer c.queue.Forget(item)

	key, err := cache.MetaNamespaceKeyFunc(item)

	if err != nil {
		fmt.Printf("Getting key from cache: %s\n", err.Error())
	}

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		fmt.Printf("Error while splitting namespace key:%s\n", err.Error())
		return false
	}

	err = c.syncDeployment(ns, name)
	if err != nil {
		fmt.Printf("syncing deployment:%s\n", err.Error())
		return false
	}

	return true
}

func (c *controller) syncDeployment(ns, name string) error {
	context := context.Background()
	//have to get deployment name to name the service
	dep, err := c.dplister.Deployments(ns).Get(name)
	if err != nil {
		fmt.Printf("Getting dep name: %s\n", err.Error())
	}

	// create a service
	//need to get pod label for service to get attached to it
	//have to specify name and stuff for the service
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dep.Name,
			Namespace: ns,
		},
		Spec: corev1.ServiceSpec{
			Selector: depLabels(*dep),
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: 80,
				},
			},
		},
	}
	s, err := c.clientset.CoreV1().Services(ns).Create(context, &svc, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("Error creating service:%s\n", err.Error())
	}
	// create ingress
	return createIngress(context, c.clientset, s)
}

func depLabels(dep appsv1.Deployment) map[string]string {
	return dep.Spec.Template.Labels
}

func createIngress(ctx context.Context, client kubernetes.Interface, svc *corev1.Service) error {
	pathType := "Prefix"
	ingress := netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/rewrite-target": "/",
			},
		},
		Spec: netv1.IngressSpec{
			Rules: []netv1.IngressRule{
				netv1.IngressRule{
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{
								netv1.HTTPIngressPath{
									Path:     fmt.Sprintf("/%s", svc.Name),
									PathType: (*netv1.PathType)(&pathType),
									Backend: netv1.IngressBackend{
										Service: &netv1.IngressServiceBackend{
											Name: svc.Name,
											Port: netv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	_, err := client.NetworkingV1().Ingresses(svc.Namespace).Create(ctx, &ingress, metav1.CreateOptions{})
	return err
}

// so whenever add and delete operations take place the events registered with the informers trigger those functions
func (c *controller) handleAdd(obj interface{}) {

	fmt.Println("Deployment reource has been added")
	c.queue.Add(obj)
}

func (c *controller) handleDel(obj interface{}) {

	fmt.Println("Deployment resource has been removed")
	c.queue.Add(obj)
}
