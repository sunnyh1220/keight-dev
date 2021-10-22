/*
Copyright 2021 sunnyh.

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

package controllers

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	etcdv1alpha1 "github.com/sunnyh1220/keight-dev/etcd-operator/api/v1alpha1"
)

// EtcdClusterReconciler reconciles a EtcdCluster object
type EtcdClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=etcd.sunnyh.easy,resources=etcdclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=etcd.sunnyh.easy,resources=etcdclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=etcd.sunnyh.easy,resources=etcdclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the EtcdCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *EtcdClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("etcdcluster", req.NamespacedName)

	// 获取EtcdCluster实例
	var etcdCluster etcdv1alpha1.EtcdCluster
	err := r.Get(ctx, req.NamespacedName, &etcdCluster)
	if err != nil {
		// EtcdCluster被删除时，忽略
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 调谐
	// CreateOrUpdate Service
	var svc corev1.Service
	svc.Name = etcdCluster.Name
	svc.Namespace = etcdCluster.Namespace
	// 冲突时重试
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		or, err := ctrl.CreateOrUpdate(ctx, r.Client, &svc, func() error {
			// 具体调谐需要在当前函数实现
			// 实际上是去拼装期望的资源对象
			MutateHeadlessService(&etcdCluster, &svc)
			return ctrl.SetControllerReference(&etcdCluster, &svc, r.Scheme)
		})
		log.Info("CreateOrUpdate", "Service", or)
		return err
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	// CreateOrUpdate StatefulSet
	var sts appsv1.StatefulSet
	sts.Name = etcdCluster.Name
	sts.Namespace = etcdCluster.Namespace
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		or, err := ctrl.CreateOrUpdate(ctx, r.Client, &sts, func() error {
			// 具体调谐需要在当前函数实现
			MutateStatefulSet(&etcdCluster, &sts)
			return ctrl.SetControllerReference(&etcdCluster, &sts, r.Scheme)
		})
		log.Info("CreateOrUpdate", "StatefulSet", or)
		return err
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EtcdClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&etcdv1alpha1.EtcdCluster{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}
