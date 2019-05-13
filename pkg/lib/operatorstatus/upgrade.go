package operatorstatus

import (
	"fmt"

	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	log "github.com/sirupsen/logrus"
)

func NewMonitor(name string, opClient operatorclient.ClientInterface, configClient configv1client.ConfigV1Interface) Monitor, error {
	return nil, nil
}

type Monitor interface {
	Run(stopCh <-chan struct{})
}


type monitor struct {

}

func (m *monitor) Run(stopCh <-chan struct{}) {
	
}





type handler struct {
}

func (h *handler) Handle(obj interface{}) error {
	clusterServiceVersion, ok := obj.(*v1alpha1.ClusterServiceVersion)

	if !ok {
		err := fmt.Errorf("wrong type - %#v", obj)
		log.Errorf("%v", err)
		return err
	}

	log.Infof("found a CSV: %s", clusterServiceVersion.GetName())

	// check if this is the right CSV.
	// now hand it over to 

	return nil
}

type 
