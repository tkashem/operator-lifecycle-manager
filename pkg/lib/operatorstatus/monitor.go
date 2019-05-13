package operatorstatus

import (
	configv1client "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/operatorclient"
)

func NewMonitor(name string, opClient operatorclient.ClientInterface, configClient configv1client.ConfigV1Interface) (Monitor, Sender) {
	return nil, nil
}

type Monitor interface {
	Run(stopCh <-chan struct{})
}

type Sender interface {
	Send()
}

type monitor struct {
}

func (m *monitor) Run(stopCh <-chan struct{}) {

}
