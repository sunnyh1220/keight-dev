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

package v1alpha1

import (
	"context"
	"regexp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource/resourcestrategy"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// User
// +k8s:openapi-gen=true
type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserSpec   `json:"spec,omitempty"`
	Status UserStatus `json:"status,omitempty"`
}

// UserList
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []User `json:"items"`
}

// UserSpec defines the desired state of User
type UserSpec struct {
	// email
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Format=email
	Email string `json:"email,omitempty"`

	// phone
	// +kubebuilder:validation:Required
	Phone string `json:"phone,omitempty"`

	// password
	// +kubebuilder:validation:Required
	Password string `json:"password,omitempty"`
}

var _ resource.Object = &User{}
var _ resourcestrategy.Validater = &User{}
var _ resourcestrategy.Defaulter = &User{}

func (in *User) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *User) NamespaceScoped() bool {
	return false
}

func (in *User) New() runtime.Object {
	return &User{}
}

func (in *User) NewList() runtime.Object {
	return &UserList{}
}

func (in *User) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "iam.keight.io",
		Version:  "v1alpha1",
		Resource: "users",
	}
}

func (in *User) IsStorageVersion() bool {
	return true
}

func (in *User) Default() {
	oldLabels := in.GetLabels()
	if oldLabels == nil {
		oldLabels = map[string]string{}
	}
	oldLabels["crateBy"] = "keight"
	in.SetLabels(oldLabels)
}

func (in *User) Validate(ctx context.Context) field.ErrorList {
	allErrs := field.ErrorList{}
	// 正则检查phone是否合法（第一位必为1的十一位数字）
	if !checkMobile(in.Spec.Phone) {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("phone"), in.Spec.Phone, "phone is invalid"))
	}

	return allErrs
}

func checkMobile(phone string) bool {
	regRuler := "^1[345789]{1}\\d{9}$"
	reg := regexp.MustCompile(regRuler)
	return reg.MatchString(phone)
}

var _ resource.ObjectList = &UserList{}

func (in *UserList) GetListMeta() *metav1.ListMeta {
	return &in.ListMeta
}

// UserStatus defines the observed state of User
type UserStatus struct {
	Phase UserPhase `json:"phase,omitempty"`
}

// UserPhase defines the observed state of User
type UserPhase string

const (
	// UserPhaseNone
	UserPhaseNone UserPhase = ""
	// UserPhasePending
	UserPhasePending UserPhase = "Pending"
	// UserPhaseActive
	UserPhaseActive UserPhase = "Active"
	// UserPhaseTerminating
	UserPhaseTerminating UserPhase = "Terminating"
)

func (in UserStatus) SubResourceName() string {
	return "status"
}

// User implements ObjectWithStatusSubResource interface.
var _ resource.ObjectWithStatusSubResource = &User{}

func (in *User) GetStatus() resource.StatusSubResource {
	return in.Status
}

// UserStatus{} implements StatusSubResource interface.
var _ resource.StatusSubResource = &UserStatus{}

func (in UserStatus) CopyTo(parent resource.ObjectWithStatusSubResource) {
	parent.(*User).Status = in
}
