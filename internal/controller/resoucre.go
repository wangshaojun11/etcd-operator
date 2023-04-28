package controller

import (
	etcdv1alpha1 "github.com/wangshaojun11/etcd-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// 定义全局标签的key
var (
	EtcdClusterLabelKey       = "etcd.uisee.com/cluster"
	EtcdClusterCommonLabelKey = "app"
	// 两处需要使用到此名字
	EtcdDataDirName = "datadir"
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
	// 定义 labels
	sts.Labels = map[string]string{
		EtcdClusterCommonLabelKey: "etcd",
	}

	// 定义sts的spec
	sts.Spec = appsv1.StatefulSetSpec{
		// 定义副本数
		Replicas: cluster.Spec.Size,
		// 定义serviceName
		ServiceName: cluster.Name, //Headless Svc name
		// 定义标签选择器
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				EtcdClusterCommonLabelKey: "etcd",
			},
		}, // 定义标签选择器结束
		// 定义 template
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					EtcdClusterCommonLabelKey: "etcd",
					EtcdClusterLabelKey:       cluster.Name,
				}, // 定义标签
			}, // 定义metadata
			Spec: corev1.PodSpec{
				Containers: newContainers(cluster), //Containers里面的东西非常多，用个函数组装，需要用到cluster里面的东西，传入cluster
			}, //定义 pod spec
		}, // 定义 template
		// 定义volume
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			corev1.PersistentVolumeClaim{ //本身是一个数组，所以是 corev1.PersistentVolumeClaim
				// 定义matadata
				ObjectMeta: metav1.ObjectMeta{
					Name: EtcdDataDirName, //因为挂载的时候使用此名字，所以定义全局变量
				},
				// 定义spec
				Spec: corev1.PersistentVolumeClaimSpec{
					// 定义 accessmodes
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadOnlyMany,
					},
					//定义 resources
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse(cluster.Spec.StorageSize), //直接写1Gi会报错，使用resource.MustParse转化格式
						},
					},
					// 定义 StorageClass
					StorageClassName: &cluster.Spec.StorageClass,
				},
			},
		},
	}
}

func newContainers(cluster *etcdv1alpha1.EtcdCluster) []corev1.Container { //Containers 的类型是 []Container 所以要返回 []Container
	panic("unimplemented")
}
