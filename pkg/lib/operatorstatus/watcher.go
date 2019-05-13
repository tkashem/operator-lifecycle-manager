package operatorstatus

import (
	"time"

	operatorsv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	operatorclient "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	client   versioned.Interface
	indexer  cache.Indexer
	informer cache.Controller
	queue    workqueue.RateLimitingInterface
	logger   *logrus.Entry
	handler  Handler
}

const (
	selectorKey     = "olm.cluster.reporting"
	selectorValue   = "true"
	reSyncPeriod    = 5 * time.Minute
	targetNamespace = v1.NamespaceAll
)

func selector() labels.Selector {
	filters := map[string]string{
		selectorKey: selectorValue,
	}

	return labels.Set(filters).AsSelector()
}

// Options specifies the parameters to configure a worker
type Options struct {
	KubeConfigPath string
	Handler        Handler
	ResyncPeriod   time.Duration
}

// New returns a controller.
func New(kubeConfigPath string, handler Handler) (controller *Controller, err error) {
	return NewWithParams(kubeConfigPath, reSyncPeriod, targetNamespace, handler, selector())
}

// NewWithParams returns a controller
func NewWithParams(kubeConfigPath string, reSyncPeriod time.Duration, namespace string, handler Handler, filter labels.Selector) (controller *Controller, err error) {
	client, err := operatorclient.NewClient(kubeConfigPath)
	if err != nil {
		return
	}

	reSyncPerod := 5 * time.Minute

	// Create a new ClusterServiceVersion watcher
	watcher := &cache.ListWatch{
		ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
			options.LabelSelector = filter.String()
			return client.OperatorsV1alpha1().ClusterServiceVersions(namespace).List(options)
		},

		WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
			options.LabelSelector = filter.String()
			return client.OperatorsV1alpha1().ClusterServiceVersions(namespace).Watch(options)
		},
	}

	// We need a queue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				return
			}

			queue.Add(key)
		},

		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err != nil {
				return
			}

			queue.Add(key)
		},

		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we
			// have to use this key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err != nil {
				return
			}

			queue.Add(key)
		},
	}

	// Bind the work queue to a cache with the help of an informer. This way we
	// make sure that whenever the cache is updated, the clusterserviceversion
	// key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might
	// see a newer version of the ClusterServiceVersion than the version which
	// was responsible for triggering the update.
	indexer, informer := cache.NewIndexerInformer(watcher, &operatorsv1alpha1.ClusterServiceVersion{}, reSyncPerod, handlers, cache.Indexers{})

	controller = &Controller{
		client:   client,
		indexer:  indexer,
		informer: informer,
		queue:    queue,
		handler:  handler,
		logger:   logrus.WithField("worker", "csv"),
	}
	return
}
