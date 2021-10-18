package controllers

import (
	"fmt"
	"github.com/sunnyh1220/keight-dev/controller-sample/pkg/generated/informers/externalversions/stable/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"time"
)

type CronTabController struct {
	informer v1alpha1.CronTabInformer
	queue    workqueue.RateLimitingInterface
}

func NewCronTabController(informer v1alpha1.CronTabInformer, queue workqueue.RateLimitingInterface) *CronTabController {
	return &CronTabController{
		informer: informer,
		queue:    queue,
	}
}

func (c *CronTabController) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Info("starting crontab controller")

	go c.informer.Informer().Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.informer.Informer().HasSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	klog.Info("shutting down pod controller")
	return nil
}

func (c *CronTabController) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *CronTabController) processNextWorkItem() bool {
	obj, quit := c.queue.Get()

	if quit {
		return false
	}

	err := func(obj interface{}) error {
		defer c.queue.Done(obj)

		key, ok := obj.(string)
		if !ok {
			c.queue.Forget(obj)
			return fmt.Errorf("expected string in workqueue but got %#v", obj)
		}

		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		c.queue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil

	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
	}

	return true
}

func (c *CronTabController) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return fmt.Errorf("invalid resource key: %s", key)
	}
	crontab, err := c.informer.Lister().CronTabs(namespace).Get(name)

	if err != nil {
		if errors.IsNotFound(err) {
			klog.Warningf("[CronTabCRD] %s/%s does not exist in local cache, will delete it from CronTab ...", namespace, name)

			klog.Infof("[CronTabCRD] deleting crontab: %s/%s ...", namespace, name)
			// todo more process detail

			return nil
		}
		utilruntime.HandleError(fmt.Errorf("failed to get crontab by: %s/%s", namespace, name))
		return err
	}

	klog.Infof("[CronTabCRD] try to process crontab: %#v ...", crontab)
	// todo more process detail

	return nil
}
