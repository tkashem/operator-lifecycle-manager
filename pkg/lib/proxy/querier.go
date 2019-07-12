package proxy

import (
	corev1 "k8s.io/api/core/v1"
)

// DefaultQuerier does...
func DefaultQuerier() Querier {
	return &defaultQuerier{}
}

// Querier is an interface that wraps the QueryProxyConfig method.
//
// QueryProxyConfig returns the global cluster level proxy env variable(s).
type Querier interface {
	QueryProxyConfig() (proxy []corev1.EnvVar, err error)
}

type defaultQuerier struct {
}

func (*defaultQuerier) QueryProxyConfig() (proxy []corev1.EnvVar, err error) {
	return
}
