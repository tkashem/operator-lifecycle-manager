package operatorstatus

import (
	configv1 "github.com/openshift/api/config/v1"
	"k8s.io/apimachinery/pkg/util/clock"
)

func newBuilder(clock clock.Clock) *builder {
	return &builder{
		clock: clock,
	}
}

type builder struct {
	status configv1.ClusterOperatorStatus
	clock  clock.Clock
}

func (b *builder) GetStatus() configv1.ClusterOperatorStatus {
	return b.status
}

func (b *builder) WithProgressing(status configv1.ConditionStatus, message string) *builder {
	return b
}

func (b *builder) WithDegraded(message string) *builder {
	return b
}

func (b *builder) WithAvailable() *builder {
	return b
}

func (b *builder) WithVersion() *builder {
	return b
}
