/*
Copyright 2023 wangshaojun.

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

package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	etcdv1alpha1 "github.com/wangshaojun11/etcd-operator/api/v1alpha1"
)

// EtcdClusterReconciler reconciles a EtcdCluster object
type EtcdClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=etcd.uisee.com,resources=etcdclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=etcd.uisee.com,resources=etcdclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=etcd.uisee.com,resources=etcdclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the EtcdCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *EtcdClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// 1. 获取 EtcdCluster 实例
	var etcdCluster etcdv1alpha1.EtcdCluster
	if err := r.Client.Get(ctx, req.NamespacedName, &etcdCluster); err != nil {
		// 返回err，把这一次调节事件放回队列中，重新处理。
		// 如果 EtcdCluster 是被删除的(NotFound)，应该被忽略.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// 已经获取到了 EtcdCluster 实例

	// 2. 创建或者更新 StatefulSet 以及 Headless SVC 对象
	// 调协：当前的状态和期望状态进行对比。CreateOrUpdate

	// 2.1 CreateOrUpdate  Service
	var svc corev1.Service
	svc.Name = etcdCluster.Name
	svc.Namespace = etcdCluster.Namespace
	// 实现调协的函数
	or, err := ctrl.CreateOrUpdate(ctx, r.Client, &svc, func() error {
		// 正在调协的函数必须在这里面实现，实际就是拼装service
		MutateHeadlessSvc(&etcdCluster, &svc)
		return controllerutil.SetControllerReference(&etcdCluster, &svc, r.Scheme)

	})
	// 表示调协出错
	if err != nil {
		return ctrl.Result{}, err
	}
	// 输出结果
	log.Info("CreateOrUpdate", "Service", or)

	// 2.2 CreateOrUpdate  Statefulset
	var sts appsv1.StatefulSet
	sts.Name = etcdCluster.Name
	sts.Namespace = etcdCluster.Namespace
	// 实现调协的函数
	or, err = ctrl.CreateOrUpdate(ctx, r.Client, &sts, func() error {
		// 正在调协的函数必须在这里面实现，实际就是拼装 Statefulset
		MutateStatefulSet(&etcdCluster, &sts)
		return controllerutil.SetControllerReference(&etcdCluster, &sts, r.Scheme)
	})
	// 表示调协出错
	if err != nil {
		return ctrl.Result{}, err
	}
	// 输出结果
	log.Info("CreateOrUpdate", "StatefulSet", or)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EtcdClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&etcdv1alpha1.EtcdCluster{}).
		Complete(r)
}
