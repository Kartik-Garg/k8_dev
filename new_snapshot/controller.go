package main

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type controller struct {
	//need clientset so we can create resources
	dynamic dynamic.Interface
	//so we can pick up things from it and perform stuff on it
	informer  cache.SharedIndexInformer
	queue     workqueue.RateLimitingInterface
	clientset kubernetes.Interface
}

func newController(dynamic dynamic.Interface, informer cache.SharedIndexInformer, clientset kubernetes.Interface) *controller {
	c := &controller{
		dynamic:   dynamic,
		informer:  informer,
		queue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "newQueue"),
		clientset: clientset,
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.createSnapshot,
	})

	return c
}

func (c *controller) run(ch <-chan struct{}) {
	fmt.Println("Starting the controller")
	if !cache.WaitForCacheSync(ch, c.informer.HasSynced) {
		fmt.Println("waiting for cache to be synced")
	}
	wait.Until(c.worker, time.Second, ch)

	<-ch
}

func (c *controller) worker() {
}

func (c *controller) createSnapshot(obj interface{}) {

	newObj := obj.(*unstructured.Unstructured)

	//fmt.Println(newObj.GetName())
	jsonObj := newObj.Object
	//fmt.Println(jsonObj)

	if jsonObj["recoverRequired"] == false {

		snap := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": fmt.Sprintf("%s/%s", "snapshot.storage.k8s.io", "v1"),
				"kind":       "VolumeSnapshot",
				"metadata": map[string]interface{}{
					"name": jsonObj["snapshotName"],
				},
				"spec": map[string]interface{}{
					"volumeSnapshotClassName": "csi-hostpath-snapclass",
					"source": map[string]interface{}{
						"persistentVolumeClaimName": jsonObj["sourcePersistentVolumeClaimName"],
					},
				},
			},
		}

		resources, err := c.dynamic.Resource(schema.GroupVersionResource{
			Group:    "snapshot.storage.k8s.io",
			Version:  "v1",
			Resource: "volumesnapshots",
		}).Namespace("snapshot").Create(context.Background(), snap, metav1.CreateOptions{})
		if err != nil {
			fmt.Printf("Error dynclient resources:%s", err.Error())
		}

		fmt.Println("\n\n\n\n", resources.GetName())

	} else {
		//to recover
		//creating pvc first
		// var storageClassName *string
		b := "csi-hostpath-sc"
		storageClassName := &b
		d := "snapshot.storage.k8s.io"
		ApiGroup := &d
		sourceSnapshotName := jsonObj["snapshotName"].(string)

		pvc, err := c.clientset.CoreV1().PersistentVolumeClaims("snapshot").Create(context.Background(), &v1.PersistentVolumeClaim{
			Spec: v1.PersistentVolumeClaimSpec{
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceName(v1.ResourceStorage): resource.MustParse("20Gi"),
					},
				},
				StorageClassName: storageClassName,
				DataSource: &v1.TypedLocalObjectReference{
					APIGroup: ApiGroup,
					Kind:     "VolumeSnapshot",
					Name:     sourceSnapshotName,
				},
				AccessModes: []v1.PersistentVolumeAccessMode{
					v1.ReadWriteMany,
				},
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: jsonObj["newPersistentVolumeClaimName"].(string),
			},
		}, metav1.CreateOptions{})

		if err != nil {
			fmt.Printf("Error creating pvc : %s", err.Error())
		}

		fmt.Printf("name of new pvc from snapshot is:%s", pvc.GetName())

		//now we also need to patch the deployment file's volume's pvc's claimname

		// data := `
		// [
		//   { "op": "replace", "path": "/spec/template/spec/volumes/persistentVolumeClaim/claimName", "value": "recoveredpv" }
		// ]
		// // `
		// data := `{"spec":{"template":{"spec":{"volumes":{"persistentVolumeClaim": {"claimName" : "mysql-pv-claim"}}}}}}`

		// updatedDep, err := c.clientset.AppsV1().Deployments("snapshot").Patch(context.Background(), "mysql", types.StrategicMergePatchType, []byte(data), metav1.PatchOptions{})
		// if err != nil {
		// 	fmt.Printf("Error while updating the deployment file:%s", err.Error())
		// }
		// fmt.Println(updatedDep.GetName())

		// //get a specific deployment
		// // specificDep, err := c.clientset.AppsV1().Deployments("snapshot").Get(context.Background(), "mysql", metav1.GetOptions{})
		// // if err != nil {
		// // 	fmt.Printf("Error while getting specific deployment: %s", err.Error())
		// // }

		// // vol := specificDep.DeepCopy()
		// // fmt.Println(vol.Spec.Template.Spec.Volumes)
	}

}
