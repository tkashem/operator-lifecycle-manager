package operatorstatus

import (
	"reflect"

	configv1 "github.com/openshift/api/config/v1"
	configv1client "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/operatorclient"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
)

func newWriter(opClient operatorclient.ClientInterface, configClient configv1client.ConfigV1Interface, reporter *reporter) *writer {
	return &writer{
		reporter:     reporter,
		opClient:     opClient,
		configClient: configClient,
	}
}

// writer encapsulates logic for cluster operator object API. It is used to
// update ClusterOperator resource.
type writer struct {
	reporter     *reporter
	opClient     operatorclient.ClientInterface
	configClient configv1client.ConfigV1Interface
}

// EnsureExists ensures that the cluster operator resource exists with a default
// status that reflects expecting status.
func (w *writer) EnsureExists(name string) (existing *configv1.ClusterOperator, err error) {
	existing, err = w.configClient.ClusterOperators().Get(name, metav1.GetOptions{})
	if err == nil {
		return
	}

	if !k8serrors.IsNotFound(err) {
		return
	}

	co := w.reporter.NewClusterOperator(name)
	created, err := w.configClient.ClusterOperators().Create(co)
	if err != nil {
		return
	}

	created.Status = co.Status
	existing, err = w.configClient.ClusterOperators().UpdateStatus(created)
	return
}

// Write updates the clusteroperator object with the new status specified.
func (w *writer) Write(context *NotificationContext) error {
	// Initially write the cluster operator object if it does not exist.
	existing, err := w.EnsureExists(context.Name)
	if err != nil {
		return err
	}

	existingStatus := existing.Status.DeepCopy()

	// Take the existing status and write the new status information on top of it.
	newStatus := w.reporter.GetExpectedStatus(existingStatus, context)
	if reflect.DeepEqual(existingStatus, newStatus) {
		return nil
	}

	existing.Status = *newStatus
	if _, err := w.configClient.ClusterOperators().UpdateStatus(existing); err != nil {
		return err
	}

	return nil
}

// IsAPIAvailable return true if cluster operator API is present on the cluster.
// Otherwise, exists is set to false.
func (w *writer) IsAPIAvailable() (exists bool, err error) {
	opStatusGV := schema.GroupVersion{
		Group:   "config.openshift.io",
		Version: "v1",
	}
	err = discovery.ServerSupportsVersion(w.opClient.KubernetesInterface().Discovery(), opStatusGV)
	if err != nil {
		return
	}

	exists = true
	return
}
