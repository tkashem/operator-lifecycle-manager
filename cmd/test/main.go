package main

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/operatorstatus"
	log "github.com/sirupsen/logrus"
)

func main() {
	stopCh := signals.SetupSignalHandler()
	config := "/home/akashem/.kube/config"

	controller, err := operatorstatus.New(config, handle)
	if err != nil {
		log.Errorf("Error - %v", err)
		return
	}

	status := controller.Run(stopCh)
	err = <-status.Error()
	if err != nil {
		log.Errorf("Error - %v", err)
		return
	}

	log.Info("Waiting for the controller to be ready")
	<-status.Ready()

	log.Info("Waiting for the controller to exit")
	<-status.Done()
}

func handle(obj interface{}) error {
	clusterServiceVersion, ok := obj.(*v1alpha1.ClusterServiceVersion)

	if !ok {
		err := fmt.Errorf("wrong type - %#v", obj)
		log.Errorf("%v", err)
		return err
	}

	log.Infof("found a CSV: %s", clusterServiceVersion.GetName())
	return nil
}
