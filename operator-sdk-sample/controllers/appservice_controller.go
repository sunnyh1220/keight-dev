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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appv1alpha1 "github.com/sunnyh1220/keight-dev/operator-sdk-sample/api/v1alpha1"
)

// AppServiceReconciler reconciles a AppService object
type AppServiceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=app.sunnyh.easy,resources=appservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.sunnyh.easy,resources=appservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.sunnyh.easy,resources=appservices/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AppService object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *AppServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	rlog := r.Log.WithValues("appservice", req.NamespacedName)

	// Reconcile successful - don't requeue
	// return ctrl.Result{}, nil
	// Reconcile failed due to error - requeue
	// return ctrl.Result{}, err
	// Requeue for any reason other than an error
	// return ctrl.Result{Requeue: true}, nil

	// 获取自定义资源实例
	var appService appv1alpha1.AppService
	err := r.Get(ctx, req.NamespacedName, &appService)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var svc corev1.Service
	svc.Name = appService.Name
	svc.Namespace = appService.Namespace
	svc.Spec = corev1.ServiceSpec{
		Type:  corev1.ServiceTypeNodePort,
		Ports: appService.Spec.Ports,
		Selector: map[string]string{
			"app": appService.Name,
		},
	}
	or, err := ctrl.CreateOrUpdate(ctx, r.Client, &svc, func() error {
		return controllerutil.SetControllerReference(&appService, &svc, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	rlog.Info("CreateOrUpdate", "Service", or)

	var deploy appsv1.Deployment
	deploy.Name = appService.Name
	deploy.Namespace = appService.Namespace
	labels := map[string]string{"app": appService.Name}
	selector := &metav1.LabelSelector{MatchLabels: labels}
	containerPorts := []corev1.ContainerPort{}
	for _, svcPort := range appService.Spec.Ports {
		cport := corev1.ContainerPort{}
		cport.ContainerPort = svcPort.TargetPort.IntVal
		containerPorts = append(containerPorts, cport)
	}
	deploySpec := appsv1.DeploymentSpec{
		Replicas: appService.Spec.Replicas,
		Selector: selector,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:            appService.Name,
						Image:           appService.Spec.Image,
						Resources:       appService.Spec.Resources,
						Ports:           containerPorts,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Env:             appService.Spec.Envs,
					},
				},
			},
		},
	}

	or, err = ctrl.CreateOrUpdate(ctx, r.Client, &deploy, func() error {
		deploy.Labels = labels
		deploy.Spec = deploySpec
		deploy.Spec.Template.Labels = labels
		return controllerutil.SetControllerReference(&appService, &deploy, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	rlog.Info("CreateOrUpdate", "Deployment", or)

	if or == controllerutil.OperationResultUpdated {
		deployKey := types.NamespacedName{Namespace: deploy.Namespace, Name: deploy.Name}
		if err = r.Client.Get(ctx, deployKey, &deploy); err != nil {
			rlog.Error(err, "unable get deployment")
			return ctrl.Result{}, err
		}

		appService.Status.DeploymentStatus = deploy.Status
		if err := r.Status().Update(ctx, &appService); err != nil {
			rlog.Error(err, "unable to update appservice status")
			return ctrl.Result{}, err
		}
		rlog.Info("update appservice status")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.AppService{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
