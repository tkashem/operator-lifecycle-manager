package operatorstatus

import (
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/csv"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

const (
	// SelectorKey is the key of the label we use to identify the
	// corresponding ClusterServiceVersion object related to the cluster operator.
	// If we want to update a cluster operator named "package-server" then the
	// corresponding ClusterServiceVersion must have the following label
	//
	// "olm.clusteroperator.name": "package-server"
	//
	SelectorKey = "olm.clusteroperator.name"
)

// NewCSVWatchNotificationHandler returns a new instance of csv.WatchNotification
// This can be used to get notification of every CSV reconciliation request.
func NewCSVWatchNotificationHandler(log *logrus.Logger, csvSet csv.SetGenerator, finder csv.ReplaceFinder, sender Sender) *handler {
	logger := log.WithField("monitor", "clusteroperator")
	return &handler{
		csvSet: csvSet,
		finder: finder,
		sender: sender,
		logger: logger,
	}
}

type handler struct {
	csvSet csv.SetGenerator
	finder csv.ReplaceFinder
	sender Sender
	logger *logrus.Entry
}

// OnAddOrUpdate is invoked when a CSV has been added or edited. We tap into
// this notification and do the following:
//
// a. Make sure this is the CSV related to the cluster operator resource we are
//    tracking. Otherwise, do nothing.
// b. If this is the right CSV then determine the new status of the cluster
//    operator object and post it to the monitor for update.
func (h *handler) OnAddOrUpdate(in *v1alpha1.ClusterServiceVersion) {
	name, matched := h.isMatchingCSV(in)
	if !matched {
		return
	}

	h.logger.Infof("OnAddOrUpdate - found a matching CSV name=%s phase=%s, sending notification", in.GetName(), in.Status.Phase)

	final := h.getLatestInReplacementChain(in)
	context := &NotificationContext{
		Current: in,
		Final:   final,
		Name:    name,
		Deleted: false,
	}

	h.sender.Send(context)
}

func (h *handler) OnDelete(in *v1alpha1.ClusterServiceVersion) {
	name, matched := h.isMatchingCSV(in)
	if !matched {
		return
	}

	h.logger.Infof("OnDelete - found a matching CSV name=%s phase=%s, sending notification", in.GetName(), in.Status.Phase)

	final := h.getLatestInReplacementChain(in)
	context := &NotificationContext{
		Current: in,
		Final:   final,
		Name:    name,
		Deleted: true,
	}

	h.sender.Send(context)
}

func (h *handler) getLatestInReplacementChain(in *v1alpha1.ClusterServiceVersion) (final *v1alpha1.ClusterServiceVersion) {
	req, _ := labels.NewRequirement(SelectorKey, selection.Exists, []string{})
	selector := labels.NewSelector().Add(*req)

	related := h.csvSet.WithNamespaceAndLabels(in.GetNamespace(), v1alpha1.CSVPhaseAny, selector)

	return h.finder.GetFinalCSVInReplacing(in, related)
}

func (h *handler) isMatchingCSV(in *v1alpha1.ClusterServiceVersion) (name string, matched bool) {
	// If it is a "copy" CSV we ignore it.
	if in.IsCopied() {
		return
	}

	// Does it have the right label?
	labels := in.GetLabels()
	if labels == nil {
		return
	}

	name, matched = labels[SelectorKey]
	return
}
