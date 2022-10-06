package main

import (
	"fmt"

	"github.com/kartik-garg/kluster/pkg/apis/kartik.dev/v1alpha1"
)

func main() {
	k := v1alpha1.Kluster{}
	fmt.Print(k)
}
