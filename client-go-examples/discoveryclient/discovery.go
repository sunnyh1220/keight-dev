package discoveryclient

import (
	"fmt"
	"github.com/sunnyh1220/keight-dev/client-go-examples/kubeconfig"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
)

func discoveryClientSample() {
	config := kubeconfig.NewConfig("C:\\Users\\zero\\GolandProjects\\github.com\\sunnyh1220\\keight-dev\\config.yaml")

	discoveryClient := discovery.NewDiscoveryClientForConfigOrDie(config)

	_, apiResourceList, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		panic(err)
	}


	for _, list := range apiResourceList {
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			panic(err)
		}
		for _, res := range list.APIResources {
			fmt.Printf("resource name: %v, group: %v, version: %v \n",res.Name, gv.Group, gv.Version)
		}

		fmt.Println("---------------------------------")
	}
}
