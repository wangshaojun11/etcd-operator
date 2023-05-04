package controller

import (
	"strconv"

	etcdv1alpha1 "github.com/wangshaojun11/etcd-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
var (
    cmd = `HOSTNAME=$(hostname)

	ETCDCTL_API=3
	
	eps() {
		EPS=""
		for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do
			EPS="${EPS}${EPS:+,}http://${SET_NAME}-${i}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2379"
		done
		echo ${EPS}
	}
	
	member_hash() {
		etcdctl member list | grep -w "$HOSTNAME" | awk '{ print $1}' | awk -F "," '{ print $1}'
	}
	
	initial_peers() {
		PEERS=""
		for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do
		  PEERS="${PEERS}${PEERS:+,}${SET_NAME}-${i}=http://${SET_NAME}-${i}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2380"
		done
		echo ${PEERS}
	}
	
	# etcd-SET_ID
	SET_ID=${HOSTNAME##*-}
	
	# adding a new member to existing cluster (assuming all initial pods are available)
	if [ "${SET_ID}" -ge ${INITIAL_CLUSTER_SIZE} ]; then
		# export ETCDCTL_ENDPOINTS=$(eps)
		# member already added?
	
		MEMBER_HASH=$(member_hash)
		if [ -n "${MEMBER_HASH}" ]; then
			# the member hash exists but for some reason etcd failed
			# as the datadir has not be created, we can remove the member
			# and retrieve new hash
			echo "Remove member ${MEMBER_HASH}"
			etcdctl --endpoints=$(eps) member remove ${MEMBER_HASH}
		fi
	
		echo "Adding new member"
	
		echo "etcdctl --endpoints=$(eps) member add ${HOSTNAME} --peer-urls=http://${HOSTNAME}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2380"
		etcdctl member --endpoints=$(eps) add ${HOSTNAME} --peer-urls=http://${HOSTNAME}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2380 | grep "^ETCD_" > /var/run/etcd/new_member_envs
	
		if [ $? -ne 0 ]; then
			echo "member add ${HOSTNAME} error."
			rm -f /var/run/etcd/new_member_envs
			exit 1
		fi
	
		echo "==> Loading env vars of existing cluster..."
		sed -ie "s/^/export /" /var/run/etcd/new_member_envs
		cat /var/run/etcd/new_member_envs
		. /var/run/etcd/new_member_envs
	
		echo "etcd --name ${HOSTNAME} --initial-advertise-peer-urls ${ETCD_INITIAL_ADVERTISE_PEER_URLS} --listen-peer-urls http://${POD_IP}:2380 --listen-client-urls http://${POD_IP}:2379,http://127.0.0.1:2379 --advertise-client-urls http://${HOSTNAME}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2379 --data-dir /var/run/etcd/default.etcd --initial-cluster ${ETCD_INITIAL_CLUSTER} --initial-cluster-state ${ETCD_INITIAL_CLUSTER_STATE}"
	
		exec etcd --listen-peer-urls http://${POD_IP}:2380 \
			--listen-client-urls http://${POD_IP}:2379,http://127.0.0.1:2379 \
			--advertise-client-urls http://${HOSTNAME}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2379 \
			--data-dir /var/run/etcd/default.etcd
	fi
	
	for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do
		while true; do
			echo "Waiting for ${SET_NAME}-${i}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local to come up"
			ping -W 1 -c 1 ${SET_NAME}-${i}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local > /dev/null && break
			sleep 1s
		done
	done
	
	echo "join member ${HOSTNAME}"
	# join member
	exec etcd --name ${HOSTNAME} \
		--initial-advertise-peer-urls http://${HOSTNAME}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2380 \
		--listen-peer-urls http://${POD_IP}:2380 \
		--listen-client-urls http://${POD_IP}:2379,http://127.0.0.1:2379 \
		--advertise-client-urls http://${HOSTNAME}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2379 \
		--initial-cluster-token etcd-cluster-1 \
		--data-dir /var/run/etcd/default.etcd \
		--initial-cluster $(initial_peers) \
		--initial-cluster-state new`
	Lifecmd = `HOSTNAME=$(hostname)

	member_hash() {
		etcdctl member list | grep -w "$HOSTNAME" | awk '{ print $1}' | awk -F "," '{ print $1}'
	}
	
	eps() {
		EPS=""
		for i in $(seq 0 $((${INITIAL_CLUSTER_SIZE} - 1))); do
			EPS="${EPS}${EPS:+,}http://${SET_NAME}-${i}.${SET_NAME}.${MY_NAMESPACE}.svc.cluster.local:2379"
		done
		echo ${EPS}
	}
	
	export ETCDCTL_ENDPOINTS=$(eps)
	SET_ID=${HOSTNAME##*-}
	
	# Removing member from cluster
	if [ "${SET_ID}" -ge ${INITIAL_CLUSTER_SIZE} ]; then
		echo "Removing ${HOSTNAME} from etcd cluster"
		etcdctl member remove $(member_hash)
		if [ $? -eq 0 ]; then
			# Remove everything otherwise the cluster will no longer scale-up
			rm -rf /var/run/etcd/*
		fi
	fi`
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
	return []corev1.Container{
		corev1.Container{ //本身是一个数组，所以是 corev1.Container
			Name:            "etcd",
			Image:           cluster.Spec.Image,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Ports: []corev1.ContainerPort{
				corev1.ContainerPort{ //本身是一个数组，所以是 corev1.ContainerPort
					Name:          "peer",
					ContainerPort: 2380,
				},
				corev1.ContainerPort{ // 数组的第二个元素，
					Name:          "client",
					ContainerPort: 2379,
				},
			},
			Env: []corev1.EnvVar{
				corev1.EnvVar{ //本身是一个数组，所以是 corev1.EnvVar
					Name:  "INITIAL_CLUSTER_SIZE",
					Value: strconv.Itoa(int(*cluster.Spec.Size)), // value的格式是string类型，转换格式
				},
				corev1.EnvVar{
					Name:  "SET_NAME",
					Value: cluster.Name,
				},
				corev1.EnvVar{
					Name: "MY_NAMESPACE",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "metadata.namespace",
						},
					},
				},
				corev1.EnvVar{
					Name: "POD_IP",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "status.podIP",
						},
					},
				},
			},
			VolumeMounts: []corev1.VolumeMount{
				corev1.VolumeMount{
					Name: EtcdDataDirName,
					MountPath: "/var/run/etcd",
				},
			},
			Command: []string{
				"/bin/sh", "-ec",
				cmd,
			},
			Lifecycle: &corev1.Lifecycle{
				PreStop: &corev1.LifecycleHandler{
					Exec: &corev1.ExecAction{
						Command: []string{
							"/bin/sh","-ec",
							Lifecmd,
						},
					},
				},
			},
		},
	}
}
