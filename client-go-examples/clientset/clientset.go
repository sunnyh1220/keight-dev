package clientset

import (
	"context"
	"fmt"
	"github.com/sunnyh1220/keight-dev/client-go-examples/kubeconfig"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func clientsetSample() {
	config := kubeconfig.NewConfig("C:\\Users\\zero\\GolandProjects\\github.com\\sunnyh1220\\keight-dev\\config.yaml")

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	pods, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{Limit: 500})
	if err != nil {
		 panic(err)
	}

	for _, item := range pods.Items {
		fmt.Println(item.Name, item.Status.Phase)
	}
}
