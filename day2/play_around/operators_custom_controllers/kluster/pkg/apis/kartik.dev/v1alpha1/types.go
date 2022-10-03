package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Kluster struct {
	//type meta and object meta can be used from apimachinery
	metav1.TypeMeta
	metav1.ObjectMeta

	//need the 3rd part called as spec meta as well.
	Spec KlusterSpec
}

type KlusterSpec struct {
	//what are all the fields that should be given to the operator
	Name    string
	Region  string
	Version string

	NodePools []NodePools
}

type NodePools struct {
	Size  string
	Name  string
	Count int
}

// need to implement list resource as well and then need to implement it as a resource as well.
type KlusterList struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Items []Kluster
}
