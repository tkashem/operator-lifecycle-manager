package operatorstatus

import (
	configv1 "github.com/openshift/api/config/v1"
)

type builder struct {
	status configv1.ClusterOperatorStatus
}

func (b *builder) GetStatus() configv1.ClusterOperatorStatus, error {
	return b.status, nil
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
