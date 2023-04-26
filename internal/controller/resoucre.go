package controller

import (
	etcdv1alpha1 "github.com/wangshaojun11/etcd-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// 定义全局标签的key
var (
	EtcdClusterLabelKey       = "etcd.uisee.com/cluster"
	EtcdClusterCommonLabelKey = "app"
)

// Service 调协实现
func MutateHeadlessSvc(cluster *etcdv1alpha1.EtcdCluster, svc *corev1.Service) {
	// 首先创建spec,(不知道svc.Spec是什么类型点击Spec)
	svc.Spec = corev1.ServiceSpec{
		ClusterIP: corev1.ClusterIPNone, // HeadlessSVC IP为 None
		Selector: map[string]string{
			EtcdClusterLabelKey: cluster.Name,
		}, // 标签要与 Sts 一致
		Ports: []corev1.ServicePort{
			corev1.ServicePort{
				Name: "client",
				Port: 2379,
			}, // client端口
			corev1.ServicePort{
				Name: "peer",
				Port: 2380,
			}, // peer 端口
		}, // 定义了两个端口
	}
	// 创建 label
	svc.Labels = map[string]string{
		EtcdClusterCommonLabelKey: "etcd",
	}
}

// StatefulSet 调协实现
func MutateStatefulSet(cluster *etcdv1alpha1.EtcdCluster, sts *appsv1.StatefulSet) {
	
}
