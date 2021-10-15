package restclient

import (
	"context"
	"fmt"
	"github.com/sunnyh1220/keight-dev/client-go-examples/kubeconfig"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

func restClientSample() {
	config := kubeconfig.NewConfig("C:\\Users\\zero\\GolandProjects\\github.com\\sunnyh1220\\keight-dev\\config.yaml")

	config.APIPath = "api"
	config.GroupVersion = &corev1.SchemeGroupVersion
	config.NegotiatedSerializer = scheme.Codecs

	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		panic(err)
	}
	result := &corev1.PodList{}
	err = restClient.Get().
		Namespace("").
		Resource("pods").
		VersionedParams(&metav1.ListOptions{Limit: 500}, scheme.ParameterCodec).
		Do(context.Background()).
		Into(result)
	if err != nil {
		panic(err)
	}
	for _, item := range result.Items {
		fmt.Println(item.Name, item.Status.Phase)
	}

}
