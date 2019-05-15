package operatorstatus

import (
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/clock"
)

func newReporter() *reporter {
	return &reporter{
		clock: &clock.RealClock{},
	}
}

// reporter provides the logic for initialzing ClusterOperator and
// ClusterOperatorStatus types.
type reporter struct {
	clock clock.Clock
}

// NewClusterOperator returns an initialized ClusterOperator object that is
// suited for creation if the given object does not exist already. The
// initialized object has the expected status for cluster operator resource
// before we have seen any corresponding CSV.
func (r *reporter) NewClusterOperator(name string) *configv1.ClusterOperator {
	builder := &Builder{
		clock: r.clock,
	}

	status := builder.WithProgressing(configv1.ConditionTrue, fmt.Sprintf("Expecting to see corresponding CSV for %s", name)).
		WithAvailable(configv1.ConditionFalse, "").
		WithDegraded(configv1.ConditionFalse).
		GetStatus()

	co := &configv1.ClusterOperator{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Status: *status,
	}

	return co
}

// GetUpdatedStatus
// Two scenarios:
// a. Fresh install of an operator (v1), no previous version installed.
//   1. Working toward v1
//   2. v1 successfully installed
//   3. v1 deploy failed
//   4. v1 has been removed, post successful install.
//
// b. Newer version of the operator (v2) is being installed
//   1. working toward v2. (v1 is being replaced, it waits until v2 successfully is successfully installed)
//      Is v1 available while v2 is being installed?
//   2.
func (r *reporter) GetExpectedStatus(existing *configv1.ClusterOperatorStatus, context *NotificationContext) *configv1.ClusterOperatorStatus {
	builder := &Builder{
		clock:  r.clock,
		status: existing,
	}

	csv := func() *v1alpha1.ClusterServiceVersion {
		if context.Final != nil {
			return context.Final
		}

		return context.Current
	}

	stale := func(message string) {
		builder.WithProgressing(configv1.ConditionFalse, message).
			WithAvailable(configv1.ConditionFalse, "").
			WithDegraded(configv1.ConditionFalse)
	}

	available := func(latest *v1alpha1.ClusterServiceVersion) {
		builder.WithProgressing(configv1.ConditionFalse, fmt.Sprintf("Deployed version %s", latest.Spec.Version)).
			WithAvailable(configv1.ConditionTrue, "").
			WithDegraded(configv1.ConditionFalse).
			WithVersion(latest.GetName(), latest.Spec.Version.String())
	}

	progressing := func(latest *v1alpha1.ClusterServiceVersion) {
		builder.WithProgressing(configv1.ConditionTrue, fmt.Sprintf("Working toward %s", latest.Spec.Version)).
			WithAvailable(configv1.ConditionFalse, "").
			WithDegraded(configv1.ConditionFalse)
	}

	// If a CSV has been deleted then the version of the deleted operator should
	// be deleted from status.
	if context.Deleted {
		current := context.Current
		builder.WithoutVersion(current.GetName(), current.Spec.Version.String())

		// If it's not a upgrade, no newer version is being installed. So we can
		// set
		if context.Final == nil {
			stale(fmt.Sprintf("Uninstalled version %s", current.Spec.Version))
			return builder.GetStatus()
		}
	}

	latest := csv()
	switch latest.Status.Phase {
	case v1alpha1.CSVPhaseSucceeded:
		available(latest)

	case v1alpha1.CSVPhaseFailed:
		stale(fmt.Sprintf("Failed to deploy %s", latest.Spec.Version))

	default:
		progressing(latest)
	}

	return builder.GetStatus()
}
