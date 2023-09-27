/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	iamv1alpha1 "github.com/sunnyh1220/keight-dev/apiserver-builder-sample/pkg/apis/iam/v1alpha1"
	"k8s.io/klog"
	"sigs.k8s.io/apiserver-runtime/pkg/builder"
)

func main() {
	err := builder.APIServer.
		// +kubebuilder:scaffold:resource-register
		WithResource(&iamv1alpha1.User{}).
		//WithResourceAndStorage(&iamv1alpha1.User{},
		//	mysql.NewMysqlStorageProvider(
		//		"mysql",
		//		int32(3306),
		//		"root",
		//		"keight",
		//		"apiserver_builder_sample",
		//	)).
		//WithResourceAndHandler(&iamv1alpha1.User{}, filepath.NewJSONFilepathStorageProvider(&iamv1alpha1.User{}, "data")).
		//WithLocalDebugExtension().
		//WithoutEtcd().
		Execute()
	if err != nil {
		klog.Fatal(err)
	}
}
