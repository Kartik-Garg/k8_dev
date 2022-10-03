package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var SchemeGroupVersion = schema.GroupVersion{
	Group:   "kartik.dev",
	Version: "v1alpha1",
}

var (
	SchemeBuilder runtime.SchemeBuilder
)

// gets called as soon as the module or the package is loaded
func init() {
	//need to call a function that is going to register our type/resource to the scheme
	SchemeBuilder.Register(addKnownTypes())
}

func addKnownTypes(scheme runtime.Scheme) error {
	//going to register the type into the scheme
	scheme.AddKnownTypes(SchemeGroupVersion, &Kluster{}, &KlusterList{})

	metav1.AddToGroupVersion(&scheme, SchemeGroupVersion)
	return nil
}
