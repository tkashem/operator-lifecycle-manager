package operatorstatus

import (
	"fmt"
	"testing"
	"time"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/clock"
)

func TestNewClusterOperator(t *testing.T) {
	fakeClock := clock.NewFakeClock(time.Now())
	reporter := &reporter{
		clock: fakeClock,
	}

	name := "foo"
	coWant := &configv1.ClusterOperator{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Status: configv1.ClusterOperatorStatus{
			Conditions: []configv1.ClusterOperatorStatusCondition{
				configv1.ClusterOperatorStatusCondition{
					Type:               configv1.OperatorProgressing,
					Status:             configv1.ConditionTrue,
					Message:            fmt.Sprintf("Expecting to see corresponding CSV for %s", name),
					LastTransitionTime: metav1.NewTime(fakeClock.Now()),
				},
				configv1.ClusterOperatorStatusCondition{
					Type:               configv1.OperatorAvailable,
					Status:             configv1.ConditionFalse,
					LastTransitionTime: metav1.NewTime(fakeClock.Now()),
				},
				configv1.ClusterOperatorStatusCondition{
					Type:               configv1.OperatorDegraded,
					Status:             configv1.ConditionFalse,
					LastTransitionTime: metav1.NewTime(fakeClock.Now()),
				},
			},
		},
	}

	coGot := reporter.NewClusterOperator(name)

	assert.Equal(t, coWant, coGot)
}
