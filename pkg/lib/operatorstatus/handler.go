package operatorstatus

import (
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/csv"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/clock"
)

const (
	SelectorKey = "olm.cluster.reporting"
)

// NewCSVWatchNotificationHandler returns
func NewCSVWatchNotificationHandler(csvSet csv.SetGenerator, replace csv.ReplaceFinder, sender Sender, clock clock.Clock) *handler {
	return &handler{
		csvSet: csvSet,
		finder: replace,
		clock:  clock,
	}
}

type handler struct {
	csvSet csv.SetGenerator
	finder csv.ReplaceFinder
	sender Sender
	clock  clock.Clock
	value  string
}

func (h *handler) OnAddOrUpdate(in *v1alpha1.ClusterServiceVersion) {
	if h.isRight(in) {
		return
	}

	selector := labels.SelectorFromSet(labels.Set{
		SelectorKey: h.value,
	})
	related := h.csvSet.WithNamespaceAndLabels(in.GetNamespace(), v1alpha1.CSVPhaseAny, selector)

	replacedBy := h.finder.IsBeingReplaced(in, related)
	if replacedBy == nil {
		newBuilder(h.clock)
	}
}

func (h *handler) OnDelete(in *v1alpha1.ClusterServiceVersion) {
	if h.isRight(in) {
		return
	}
}

func (h *handler) isRight(in *v1alpha1.ClusterServiceVersion) bool {
	// If it is a "copy" CSV we ignore it.
	if in.IsCopied() {
		return false
	}

	// Does it have the right label?
	labels := in.GetLabels()
	if labels == nil {
		return false
	}

	value, exists := labels[SelectorKey]
	if !exists {
		return false
	}

	if value != h.value {
		return false
	}

	return true
}
