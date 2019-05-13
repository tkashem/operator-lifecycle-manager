package csv

import (
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewReplaceFinder(logger *logrus.Logger, client versioned.Interface) ReplaceFinder {
	return &replace{
		logger: logger,
		client: client,
	}
}

type ReplaceFinder interface {
	IsBeingReplaced(in *v1alpha1.ClusterServiceVersion, csvsInNamespace map[string]*v1alpha1.ClusterServiceVersion) (replacedBy *v1alpha1.ClusterServiceVersion)
	IsReplacing(in *v1alpha1.ClusterServiceVersion) *v1alpha1.ClusterServiceVersion
}

type replace struct {
	logger *logrus.Logger
	client versioned.Interface
}

func (r *replace) IsBeingReplaced(in *v1alpha1.ClusterServiceVersion, csvsInNamespace map[string]*v1alpha1.ClusterServiceVersion) (replacedBy *v1alpha1.ClusterServiceVersion) {
	for _, csv := range csvsInNamespace {
		r.logger.Infof("checking %s", csv.GetName())
		if csv.Spec.Replaces == in.GetName() {
			r.logger.Infof("%s replaced by %s", in.GetName(), csv.GetName())
			replacedBy = csv.DeepCopy()
			return
		}
	}

	return
}

func (r *replace) IsReplacing(in *v1alpha1.ClusterServiceVersion) *v1alpha1.ClusterServiceVersion {
	r.logger.Debugf("checking if csv is replacing an older version")
	if in.Spec.Replaces == "" {
		return nil
	}

	// using the client instead of a lister; missing an object because of a cache sync can cause upgrades to fail
	previous, err := r.client.OperatorsV1alpha1().ClusterServiceVersions(in.GetNamespace()).Get(in.Spec.Replaces, metav1.GetOptions{})
	if err != nil {
		r.logger.WithField("replacing", in.Spec.Replaces).WithError(err).Debugf("unable to get previous csv")
		return nil
	}

	return previous
}
