package csv

import (
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
)

type WatchNotification interface {
	OnAddOrUpdate(in *v1alpha1.ClusterServiceVersion)
	OnDelete(in *v1alpha1.ClusterServiceVersion)
}
