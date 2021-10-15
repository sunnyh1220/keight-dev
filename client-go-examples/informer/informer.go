package informer

import (
	"fmt"
	"github.com/sunnyh1220/keight-dev/client-go-examples/kubeconfig"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"time"
)

func informerSample() {
	config := kubeconfig.NewConfig("C:\\Users\\zero\\GolandProjects\\github.com\\sunnyh1220\\keight-dev\\config.yaml")

	clientset := kubernetes.NewForConfigOrDie(config)
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Minute)

	podInformer := informerFactory.Core().V1().Pods()

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			fmt.Println("add a pod:", pod.Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldPod := oldObj.(*corev1.Pod)
			newPod := newObj.(*corev1.Pod)
			fmt.Println("update pod:", oldPod.Name, newPod.Name)
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			fmt.Println("delete a pod:", pod.Name)
		},
	})

	stopper := make(chan struct{})
	defer close(stopper)

	informerFactory.Start(stopper)

	informerFactory.WaitForCacheSync(stopper)

	pods, err := podInformer.Lister().Pods("").List(labels.Everything())
	if err != nil {
		panic(err)
	}

	for _, pod := range pods {
		fmt.Println(pod.Name, pod.Spec.NodeName)
	}

	<- stopper


}
