package operatorstatus

import (
	"reflect"

	configv1 "github.com/openshift/api/config/v1"
	configv1client "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/operatorclient"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
)

type monitor interface {
	Receiver() <-chan error
}

type writer struct {
	opClient     operatorclient.ClientInterface
	configClient configv1client.ConfigV1Interface
	monitor      monitor
	name         string
}

func (w *writer) Write(newStatus configv1.ClusterOperatorStatus) error {
	// Initially write the cluster operator object if it does not exist.
	existing, err := w.configClient.ClusterOperators().Get(name, metav1.GetOptions{})
	if err != nil {
		if !k8ek8serrors.IsNotFound(err) {
			return
		}

		created, createErr := w.configClient.ClusterOperators().Create()
		if createErr != nil {
			return createErr
		}

		existing = created
	}

	existingStatus := existing.Status.DeepCopy()
	if reflect.DeepEqual(existingStatus, newStatus) {
		// log and return
		return nil
	}

	if err := w.configClient.ClusterOperators().UpdateStatus(newStatus); err != nil {
		return err
	}

	return err
}

func (w *writer) IsAPIAvailable() (exists bool, err error) {
	opStatusGV := schema.GroupVersion{
		Group:   "config.openshift.io",
		Version: "v1",
	}
	err = discovery.ServerSupportsVersion(w.opClient.KubernetesInterface().Discovery(), opStatusGV)
	if err != nil {
		log.Infof("ClusterOperator api not present, skipping update (%v)", err)
		return
	}

	exists = true
	return
}
