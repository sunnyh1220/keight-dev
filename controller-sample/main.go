package main

import (
	"flag"
	"github.com/sunnyh1220/keight-dev/controller-sample/controllers"
	"github.com/sunnyh1220/keight-dev/controller-sample/pkg/apis/stable/v1alpha1"
	clientset "github.com/sunnyh1220/keight-dev/controller-sample/pkg/generated/clientset/versioned"
	informers "github.com/sunnyh1220/keight-dev/controller-sample/pkg/generated/informers/externalversions"
	"github.com/sunnyh1220/keight-dev/controller-sample/pkg/signals"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"path/filepath"
	"time"
)

func main() {
	config := newConfig("")
	clientset, err := clientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Minute)
	crontabInformer := informerFactory.Stable().V1alpha1().CronTabs()

	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "crontab-controller")

	crontabInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				utilruntime.HandleError(err)
			}
			queue.AddRateLimited(key)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldCronTab := oldObj.(*v1alpha1.CronTab)
			newCronTab := newObj.(*v1alpha1.CronTab)
			if oldCronTab.ResourceVersion == newCronTab.ResourceVersion {
				return
			}
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			if err != nil {
				utilruntime.HandleError(err)
			}
			queue.AddRateLimited(key)
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err != nil {
				utilruntime.HandleError(err)
			}
			queue.AddRateLimited(key)
		},
	})

	cronTabController := controllers.NewCronTabController(crontabInformer, queue)

	stopCh := signals.SetupSignalHandler()

	if err := cronTabController.Run(1, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}

// path: absolute path to the kubeconfig file
func newConfig(path string) *rest.Config {
	var err error
	var config *rest.Config
	var kubeconfig *string

	if path != "" {
		kubeconfig = flag.String("kubeconfig", path, "absolute path to the kubeconfig file")
	} else if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	if config, err = rest.InClusterConfig(); err != nil {
		if config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig); err != nil {
			panic(err.Error())
		}
	}

	return config
}
