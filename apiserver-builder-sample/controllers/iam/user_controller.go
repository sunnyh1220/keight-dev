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

package iam

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	iamv1alpha1 "github.com/sunnyh1220/keight-dev/apiserver-builder-sample/pkg/apis/iam/v1alpha1"
)

// UserReconciler reconciles a User object
type UserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=iam,resources=users,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=iam,resources=users/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=iam,resources=users/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the User object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *UserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling User")

	instance := &iamv1alpha1.User{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		logger.Error(err, "unable to fetch User")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// create pod
	pod := corev1.Pod{}
	pod.Name = instance.Name
	pod.Namespace = corev1.NamespaceDefault
	pod.Spec.Containers = []corev1.Container{
		{
			Name:  "nginx",
			Image: "nginx:latest",
		},
	}

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, &pod, func() error {
		pod.Labels = instance.Labels
		return ctrl.SetControllerReference(instance, &pod, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	// update status
	instance.Status.Phase = iamv1alpha1.UserPhasePending

	// wait for pod running
	podKey := client.ObjectKey{
		Namespace: pod.Namespace,
		Name:      pod.Name,
	}
	err = r.Get(ctx, podKey, &pod)
	if err != nil {
		logger.Error(err, "unable to fetch Pod")
		return ctrl.Result{}, err
	}
	if pod.Status.Phase != corev1.PodRunning {
		return ctrl.Result{RequeueAfter: time.Second * 3}, nil
	}

	// update status
	instance.Status.Phase = iamv1alpha1.UserPhaseActive
	err = r.Status().Update(ctx, instance)
	if err != nil {
		logger.Error(err, "unable to update User status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *UserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&iamv1alpha1.User{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
