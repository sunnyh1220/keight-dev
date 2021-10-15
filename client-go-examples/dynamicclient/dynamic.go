package dynamicclient

import (
	"context"
	"fmt"
	"github.com/sunnyh1220/keight-dev/client-go-examples/kubeconfig"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func dynamicClientSample() {
	config := kubeconfig.NewConfig("C:\\Users\\zero\\GolandProjects\\github.com\\sunnyh1220\\keight-dev\\config.yaml")

	dynamicClient := dynamic.NewForConfigOrDie(config)

	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	unstructObj, err := dynamicClient.Resource(gvr).Namespace("").List(context.Background(), metav1.ListOptions{
		Limit: 500,
	})
	if err != nil {
		panic(err)
	}

	//fmt.Println(unstructObj.UnstructuredContent())

	pods := &corev1.PodList{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructObj.UnstructuredContent(), pods)
	if err != nil {
		panic(err)
	}

	for _, item := range pods.Items {
		fmt.Println(item.Name, item.Status.Phase)
	}
}
