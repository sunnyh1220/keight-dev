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
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	etcdv1alpha1 "github.com/sunnyh1220/keight-dev/etcd-operator/api/v1alpha1"
)

// EtcdBackupReconciler reconciles a EtcdBackup object
type EtcdBackupReconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	Recorder    record.EventRecorder
	BackupImage string
}

//+kubebuilder:rbac:groups=etcd.sunnyh.easy,resources=etcdbackups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=etcd.sunnyh.easy,resources=etcdbackups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=etcd.sunnyh.easy,resources=etcdbackups/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the EtcdBackup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *EtcdBackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("etcdbackup", req.NamespacedName)

	// 获取backup state
	state, err := r.getState(ctx, req)
	if err != nil {
		return ctrl.Result{}, err
	}

	// 根据状态执行相应的action
	var action Action

	// 开始判断状态
	switch {
	case state.backup == nil: // 被删除了
		log.Info("Backup Object not found")
	case !state.backup.DeletionTimestamp.IsZero(): // 被标记为删除了
		log.Info("Backup Object has been deleted")
	case state.backup.Status.Phase == "": // 要开始备份了，先标记状态为备份中
		log.Info("Backup starting...")
		newBackup := state.backup.DeepCopy()
		newBackup.Status.Phase = etcdv1alpha1.EtcdBackupPhaseBackingUp // 更改成备份中...
		action = &PatchStatus{client: r.Client, original: state.backup, new: newBackup}
	case state.backup.Status.Phase == etcdv1alpha1.EtcdBackupPhaseFailed: // 失败了
		log.Info("Backup has failed. Ignoring...")
	case state.backup.Status.Phase == etcdv1alpha1.EtcdBackupPhaseCompleted: // 完成了
		log.Info("Backup has completed. Ignoring...")
	case state.actual.pod == nil: // 当前还没有执行任务的Pod
		log.Info("Backup Pod does not exists. Creating...")
		action = &CreateObject{client: r.Client, obj: state.desired.pod}
		r.Recorder.Event(state.backup, corev1.EventTypeNormal, "SuccessfulCreate", fmt.Sprintf("Created Pod: %s", state.desired.pod.Name))
	case state.actual.pod.Status.Phase == corev1.PodFailed: // Pod执行失败
		log.Info("Backup Pod failed.")
		newBackup := state.backup.DeepCopy()
		newBackup.Status.Phase = etcdv1alpha1.EtcdBackupPhaseFailed // 更改成备份失败
		action = &PatchStatus{client: r.Client, original: state.backup, new: newBackup}
		r.Recorder.Event(state.backup, corev1.EventTypeWarning, "BackupFailed", "Backup failed. See backup pod logs for details.")
	case state.actual.pod.Status.Phase == corev1.PodSucceeded: // Pod 执行成功
		log.Info("Backup Pod success.")
		newBackup := state.backup.DeepCopy()
		newBackup.Status.Phase = etcdv1alpha1.EtcdBackupPhaseCompleted // 更改成备份成功
		action = &PatchStatus{client: r.Client, original: state.backup, new: newBackup}
		r.Recorder.Event(state.backup, corev1.EventTypeNormal, "BackupSucceeded", "Backup completed successfully")
	}

	if action != nil {
		if err := action.Execute(ctx); err != nil {
			return ctrl.Result{}, fmt.Errorf("executing action error: %s", err)
		}
	}

	return ctrl.Result{}, nil
}

// 包含EtcdBackup真实状态和期望状态
type backupState struct {
	// EtcdBackup资源对象本身
	backup *etcdv1alpha1.EtcdBackup
	// 真实的状态
	actual *backupStateContainer
	// 期望的状态
	desired *backupStateContainer
}

// backupStateContainer 包含 EtcdBackup 的状态
type backupStateContainer struct {
	pod *corev1.Pod
}

// 设置真实的状态
func (r *EtcdBackupReconciler) setActualState(ctx context.Context, state *backupState) error {
	var actual backupStateContainer

	key := client.ObjectKey{
		Namespace: state.backup.Namespace,
		Name:      state.backup.Name,
	}

	// 获取pod
	actual.pod = &corev1.Pod{}
	if err := r.Get(ctx, key, actual.pod); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("getting pod error: %s", err)
		}
		actual.pod = nil
	}

	state.actual = &actual
	return nil
}

// 设置期望的状态
func (r *EtcdBackupReconciler) setDesiredState(ctx context.Context, state *backupState) error {
	var desired backupStateContainer

	// 创建pod
	pod, err := podForBackup(state.backup, r.BackupImage)
	if err != nil {
		return fmt.Errorf("computing pod for backup error: %q", err)
	}

	// controller reference
	if err = ctrl.SetControllerReference(state.backup, pod, r.Scheme); err != nil {
		return fmt.Errorf("setting pod controller reference error : %s", err)
	}

	desired.pod = pod

	state.desired = &desired
	return nil
}

// 获取整个状态
func (r EtcdBackupReconciler) getState(ctx context.Context, req ctrl.Request) (*backupState, error) {
	var state backupState

	state.backup = &etcdv1alpha1.EtcdBackup{}
	if err := r.Get(ctx, req.NamespacedName, state.backup); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return nil, fmt.Errorf("getting backup error: %s", err)
		}
		// 已经删除
		state.backup = nil
		return &state, nil
	}

	// 获取真实状态
	if err := r.setActualState(ctx, &state); err != nil {
		return nil, fmt.Errorf("setting actual state error: %s", err)
	}

	// 获取期望状态
	if err := r.setDesiredState(ctx, &state); err != nil {
		return nil, fmt.Errorf("setting desired state error: %s", err)
	}

	return &state, nil
}

// 创建pod运行备份任务
func podForBackup(backup *etcdv1alpha1.EtcdBackup, image string) (*corev1.Pod, error) {
	var secretRef *corev1.SecretEnvSource
	var backupURL, backupEndpoint string

	if backup.Spec.StorageType == etcdv1alpha1.EtcdBackupStorageTypeS3 {
		backupURL = fmt.Sprintf("%s://%s", backup.Spec.StorageType, backup.Spec.S3.Path)
		backupEndpoint = backup.Spec.S3.Endpoint
		secretRef = &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: backup.Spec.S3.Secret,
			},
		}
	} else {
		backupURL = fmt.Sprintf("%s://%s", backup.Spec.StorageType, backup.Spec.OSS.Path)
		backupEndpoint = backup.Spec.OSS.Endpoint
		secretRef = &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: backup.Spec.OSS.Secret,
			},
		}
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: backup.Namespace,
			Name:      backup.Name,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            "backup-agent",
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Args: []string{
						"--etcd-url", backup.Spec.EtcdUrl,
						"--backup-url", backupURL,
					},
					Env: []corev1.EnvVar{
						{
							Name:  "ENDPOINT",
							Value: backupEndpoint,
						},
					},
					EnvFrom: []corev1.EnvFromSource{
						{
							SecretRef: secretRef,
						},
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("50Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("50Mi"),
						},
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EtcdBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&etcdv1alpha1.EtcdBackup{}).
		Complete(r)
}
