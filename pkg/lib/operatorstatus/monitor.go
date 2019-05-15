package operatorstatus

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	configv1client "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/operatorclient"
	"github.com/sirupsen/logrus"
)

const (
	// Wait time before we probe next while checking whether cluster
	// operator API is available.
	defaultProbeInterval = 1 * time.Minute

	// Default size of the notification channel.
	defaultNotificationChannelSize = 64
)

// NewMonitor returns a new instance of Monitor that can be used to continuously
// update a clusteroperator resource and an instance of Sender that can be used
// to send update notifications to it.
//
// The name of the clusteroperator resource to monitor is specified in name.
func NewMonitor(name string, log *logrus.Logger, opClient operatorclient.ClientInterface, configClient configv1client.ConfigV1Interface) (Monitor, Sender) {
	logger := log.WithField("monitor", "clusteroperator")
	reporter := newReporter()

	names := split(name)
	logger.Infof("monitoring the following components %s", names)

	monitor := &monitor{
		logger:         logger,
		writer:         newWriter(opClient, configClient, reporter),
		notificationCh: make(chan *NotificationContext, defaultNotificationChannelSize),
		names:          names,
	}

	return monitor, monitor
}

// NotificationContext contains all necessary information related to a notification.
type NotificationContext struct {
	Name    string
	Current *v1alpha1.ClusterServiceVersion
	Final   *v1alpha1.ClusterServiceVersion
	Deleted bool
}

// Monitor is an interface that wraps the Run method.
//
// Run is a continuous loop reads from an underlying notification channel and
// updates an clusteroperator resource.
// If the specified stop channel is closed the loop must terminate gracefully.
type Monitor interface {
	Run(stopCh <-chan struct{})
}

// Sender is an interface that wraps the Send method.
//
// Send can be used to send notification(s) to the underlying monitor. Send is a
// non-blocking operation.
// If the underlying monitor is not ready to accept the notification will be lost.
// If the status specified is nil then it is ignored.
type Sender interface {
	Send(context *NotificationContext)
}

func (n *NotificationContext) String() string {
	replaces := "<nil>"
	if n.Final != nil {
		replaces = n.Final.GetName()
	}

	return fmt.Sprintf("name=%s csv=%s deleted=%s replaces=%s", n.Name, n.Current.GetName(), strconv.FormatBool(n.Deleted), replaces)
}

type monitor struct {
	notificationCh chan *NotificationContext
	writer         *writer
	logger         *logrus.Entry
	names          []string
}

func (m *monitor) Send(context *NotificationContext) {
	if context == nil {
		return
	}

	select {
	case m.notificationCh <- context:
	default:
		m.logger.Warn("monitor not ready to accept cluster operator update notification")
	}
}

func (m *monitor) Run(stopCh <-chan struct{}) {
	m.logger.Info("starting clusteroperator monitor loop")
	defer func() {
		m.logger.Info("exiting from clusteroperator monitor loop")
	}()

	// First, we need to ensure that cluster operator API is available.
	// We will keep probing until it is available.
	for {
		exists, err := m.writer.IsAPIAvailable()
		if err != nil {
			m.logger.Infof("ClusterOperator api not present, skipping update (%v)", err)
		}

		if exists {
			m.logger.Info("ClusterOperator api is present")
			break
		}

		// Wait before next probe, or quit if parent has asked to do so.
		select {
		case <-time.After(defaultProbeInterval):
		case <-stopCh:
			return
		}
	}

	// If we are here, cluster operator is available.
	// We are expecting CSV notification which may never arrive. So the safe
	// thing to do here is write an initial ClusterOperator object with an
	// expectation.
	m.logger.Info("ensuring that all clusteroperator resources exist")
	for _, name := range m.names {
		if _, err := m.writer.EnsureExists(name); err != nil {
			m.logger.Errorf("failed to write initial clusteroperator name=%s - %v", name, err)
			break
		}
	}

	for {
		select {
		case context := <-m.notificationCh:
			if context != nil {
				m.logger.Infof("notification %s", context)
				if err := m.writer.Write(context); err != nil {
					m.logger.Errorf("failed to update, clusteroperator=%s - %v", context.Name, err)
				}
			}

		case <-stopCh:
			return
		}
	}
}

func split(n string) []string {
	names := make([]string, 0)

	values := strings.Split(n, ",")
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			names = append(names, v)
		}
	}

	return names
}
