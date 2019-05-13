package operatorstatus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	log "github.com/sirupsen/logrus"
)

func Test(t *testing.T) {
	stopCh := signals.SetupSignalHandler()
	config := "/home/akashem/.kube/config"

	controller, err := New(config, handle)
	require.NoError(t, err)

	status := controller.Run(stopCh)
	errGot := <-status.Error()
	assert.NoError(t, errGot)

	log.Info("Waiting for the controller to be ready")
	<-status.Ready()

	log.Info("Waiting for the controller to exit")
	<-status.Done()
}

func handle(key interface{}) error {
	return nil
}
