package operatorstatus

import (
	"fmt"

	cache "k8s.io/client-go/tools/cache"
)

// WorkerStatus returns
type WorkerStatus interface {
	Done() <-chan struct{}
	Ready() <-chan struct{}
	Error() <-chan error
}

// Handler is the function that reconciles the controlled object when seen
type Handler func(obj interface{}) error

// Run runs
func (c *Controller) Run(stopCh <-chan struct{}) WorkerStatus {
	result := &result{
		ready:   make(chan struct{}),
		done:    make(chan struct{}),
		errChan: make(chan error),
	}

	setup := func() {
		// Let the worker stop when we are done.
		defer func() {
			c.queue.ShutDown()

			close(result.ready)
			close(result.done)
			c.logger.Info("Run function done")
		}()

		c.logger.Info("Starting informer")
		go c.informer.Run(stopCh)

		// Wait for all involved caches to be synced, before processing items from the queue is started
		c.logger.Info("Waiting for caches to be synced")
		if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
			c.logger.Errorf("Timed out waiting for caches to sync")
			result.errChan <- fmt.Errorf("Timed out waiting for caches to sync")
			return
		}

		c.logger.Info("starting worker")
		go c.runWorker()

		result.errChan <- nil
		result.ready <- struct{}{}

		c.logger.Info("ready, now waiting for stop channel to close")
		<-stopCh
	}

	go setup()
	return result
}

func (c *Controller) runWorker() {
	for c.processNextItem() {

	}
}

func (c *Controller) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}

	defer c.queue.Done(key)

	obj, exists, err := c.indexer.GetByKey(key.(string))
	if err != nil {
		c.logger.Infof("Fetching object with key %s from store failed with %v", key, err)
		return true
	}

	if !exists {
		c.logger.Infof("The object with key[%s] does not exist anymore", key)
		return true
	}

	err = c.handler(obj)
	if err == nil {
		c.queue.Forget(key)
		return true
	}

	if c.queue.NumRequeues(key) < 5 {
		c.queue.AddRateLimited(key)
		return true
	}

	c.queue.Forget(key)
	return true
}

type result struct {
	ready   chan struct{}
	done    chan struct{}
	errChan chan error
}

func (r *result) Done() <-chan struct{} {
	return r.done
}

func (r *result) Ready() <-chan struct{} {
	return r.ready
}

func (r *result) Error() <-chan error {
	return r.errChan
}
