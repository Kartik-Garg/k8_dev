package main

import (
	"context"
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-driver-host-path/pkg/hostpath"
)

func getCSIDriver() {
	csiDriver, err := hostpath.NewHostPathDriver(hostpath.Config{
		DriverName: "goDriver",
		NodeID:     "kind-control-plane",
		Endpoint:   "tcp://127.0.0.1:10000",
		StateDir:   "/home/kartik/Documents",
	})
	if err != nil {
		fmt.Printf("Getting error while creating csi driver: %s", err.Error())
	}
	fmt.Println(csiDriver)

	//creating s snapshot
	volumeSnapshot, err := csiDriver.CreateSnapshot(context.Background(), &csi.CreateSnapshotRequest{
		SourceVolumeId: "pvc-0ccc4086-c661-4e8b-a545-977b0be33830",
		Name:           "golangdriver",
	})
	if err != nil {
		fmt.Printf("Getting error while creating snapshot: %s", err.Error())
	}
	fmt.Println(volumeSnapshot)
}
