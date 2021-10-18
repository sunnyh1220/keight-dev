package workqueue

import (
	"fmt"
	"github.com/sunnyh1220/keight-dev/client-go-examples/kubeconfig"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	informerscorev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"time"
)

func workqueueSample() {
	config := kubeconfig.NewConfig("C:\\Users\\zero\\GolandProjects\\github.com\\sunnyh1220\\keight-dev\\config.yaml")

	clientset := kubernetes.NewForConfigOrDie(config)
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Minute)

	podInformer := informerFactory.Core().V1().Pods()

	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pod-controller")

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.AddRateLimited(key)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			if err == nil {
				queue.AddRateLimited(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.AddRateLimited(key)
			}
		},
	})

	controller := NewPodController(podInformer, queue)

	// start controller
	stopCh := make(chan struct{})
	defer close(stopCh)
	go controller.Run(1, stopCh)

	select {}

}

type PodController struct {
	podInformer informerscorev1.PodInformer
	queue       workqueue.RateLimitingInterface
}

func (c *PodController) Run(workers int, stopCh chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Info("starting pod controller")

	go c.podInformer.Informer().Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.podInformer.Informer().HasSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	klog.Info("shutting down pod controller")
}

func (c *PodController) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *PodController) processNextWorkItem() bool {
	key, quit := c.queue.Get()

	if quit {
		return false
	}

	defer c.queue.Done(key)

	// 业务逻辑
	err := c.syncToStdout(key.(string))

	c.handleErr(err, key)

	return true
}

func (c *PodController) syncToStdout(key string) error {
	// 从本地存储Indexer中取出key对应的对象
	obj, exists, err := c.podInformer.Informer().GetIndexer().GetByKey(key)

	if err != nil {
		klog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		fmt.Printf("Pod %s does not exist anymore\n", key)
	} else {
		fmt.Printf("Sync/Add/Update for Pod %s\n", obj.(*corev1.Pod).GetName())
		// todo 业务逻辑
	}
	return nil
}

func (c *PodController) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	// 如果出现问题，此控制器将重试5次
	if c.queue.NumRequeues(key) < 5 {
		klog.Infof("Error syncing pod %v: %v", key, err)
		// 重新加入 key 到限速队列
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)

	// 多次重试也不能处理该key
	utilruntime.HandleError(err)
	klog.Infof("Dropping pod %q out of the queue: %v", key, err)
}

func NewPodController(informer informerscorev1.PodInformer, queue workqueue.RateLimitingInterface) *PodController {
	return &PodController{
		podInformer: informer,
		queue:       queue,
	}
}
