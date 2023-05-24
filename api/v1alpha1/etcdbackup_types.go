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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	EtcdBackPhaseBackingUp EtcdBackPhase = "BackingUp" //备份中
	EtcdBackPhaseCompleted EtcdBackPhase = "Completed" //备份完成
	EtcdBackPhaseFailed    EtcdBackPhase = "Failed"    //备份失败
)

type BackupStorageType string //定义 StorageType 的类型
type EtcdBackPhase string     // 定义 etcd 备份状态

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EtcdBackupSpec defines the desired state of EtcdBackup
type EtcdBackupSpec struct { // EtcdBackupSpec 表示 Spec 下面的内容
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	EtcdUrl      []string          `json:"etcdUrl"`
	StorageType  BackupStorageType `json:"storageType"`
	BackupSource `json:",inline"`  // 内链, 需要定义一个结构体同名

}

type BackupSource struct { //备份存储的路径。S3和OSS可选。下面分别定义S3和OSS的结构体
	S3  *S3BackupSource  `json:"s3,omitempty"`
	OSS *OSSBackupSource `json:"oss,omitempty"`
}

type S3BackupSource struct { //S3结构体
	Path     string `json:"path"`
	S3Secret string `json:"s3Secret"`
}

type OSSBackupSource struct { // OSS结构体
	Path      string `json:"path"`
	OSSSecret string `json:"ossSecret"`
}

// EtcdBackupStatus defines the observed state of EtcdBackup
type EtcdBackupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Phase          EtcdBackPhase `json:"phase,omitempty"`          // 备份状态
	StartTime      *metav1.Time  `json:"startTime,omitempty"`      // 开始备份时间
	CompletionTime *metav1.Time  `json:"completionTime,omitempty"` // 结束备份时间

}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// EtcdBackup is the Schema for the etcdbackups API
type EtcdBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EtcdBackupSpec   `json:"spec,omitempty"`
	Status EtcdBackupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EtcdBackupList contains a list of EtcdBackup
type EtcdBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EtcdBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EtcdBackup{}, &EtcdBackupList{})
}
