package interfaces

import (
	cfgappsv1 "github.com/zxq-bit/kube-admission-test/pkg/admission/framework/interfaces/apis/apps/v1"
	cfgcorev1 "github.com/zxq-bit/kube-admission-test/pkg/admission/framework/interfaces/apis/core/v1"
)

type Model interface {
	Name() string
	Start(stopCh <-chan struct{})
}

type ModelCollection struct {
	ModelMap map[string]Model
}

type ConfigCollection struct {
	DaemonSetConfig             cfgappsv1.DaemonSetConfig
	DeploymentConfig            cfgappsv1.DeploymentConfig
	ReplicaSetConfig            cfgappsv1.ReplicaSetConfig
	StatefulSetConfig           cfgappsv1.StatefulSetConfig
	ConfigMapConfig             cfgcorev1.ConfigMapConfig
	PodConfig                   cfgcorev1.PodConfig
	SecretConfig                cfgcorev1.SecretConfig
	PersistentVolumeConfig      cfgcorev1.PersistentVolumeConfig
	PersistentVolumeClaimConfig cfgcorev1.PersistentVolumeClaimConfig
}
