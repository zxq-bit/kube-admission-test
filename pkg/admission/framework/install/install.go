package install

import (
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"

	cfgappsv1 "github.com/zxq-bit/kube-admission-test/pkg/admission/framework/configs/apps/v1"
	cfgcorev1 "github.com/zxq-bit/kube-admission-test/pkg/admission/framework/configs/core/v1"
)

var (
	DaemonSetConfig             cfgappsv1.DaemonSetConfig
	DeploymentConfig            cfgappsv1.DeploymentConfig
	ReplicaSetConfig            cfgappsv1.ReplicaSetConfig
	StatefulSetConfig           cfgappsv1.StatefulSetConfig
	ConfigMapConfig             cfgcorev1.ConfigMapConfig
	PodConfig                   cfgcorev1.PodConfig
	SecretConfig                cfgcorev1.SecretConfig
	PersistentVolumeConfig      cfgcorev1.PersistentVolumeConfig
	PersistentVolumeClaimConfig cfgcorev1.PersistentVolumeClaimConfig
)

func init() {
	DaemonSetConfig.Register(arv1b1.Create)
	DaemonSetConfig.Register(arv1b1.Update)
	DaemonSetConfig.Register(arv1b1.Delete)

	DeploymentConfig.Register(arv1b1.Create)
	DeploymentConfig.Register(arv1b1.Update)
	DeploymentConfig.Register(arv1b1.Delete)

	ReplicaSetConfig.Register(arv1b1.Create)
	ReplicaSetConfig.Register(arv1b1.Update)
	ReplicaSetConfig.Register(arv1b1.Delete)

	StatefulSetConfig.Register(arv1b1.Create)
	StatefulSetConfig.Register(arv1b1.Update)
	StatefulSetConfig.Register(arv1b1.Delete)

	ConfigMapConfig.Register(arv1b1.Create)
	ConfigMapConfig.Register(arv1b1.Update)
	ConfigMapConfig.Register(arv1b1.Delete)

	PodConfig.Register(arv1b1.Create)
	PodConfig.Register(arv1b1.Update)
	PodConfig.Register(arv1b1.Delete)

	SecretConfig.Register(arv1b1.Create)
	SecretConfig.Register(arv1b1.Update)
	SecretConfig.Register(arv1b1.Delete)

	PersistentVolumeConfig.Register(arv1b1.Create)
	PersistentVolumeConfig.Register(arv1b1.Update)
	PersistentVolumeConfig.Register(arv1b1.Delete)

	PersistentVolumeClaimConfig.Register(arv1b1.Create)
	PersistentVolumeClaimConfig.Register(arv1b1.Update)
	PersistentVolumeClaimConfig.Register(arv1b1.Delete)
}

func GetConfigs() []processor.Config {
	raw := []*processor.Config{
		DaemonSetConfig.ToConfig(),
		DeploymentConfig.ToConfig(),
		ReplicaSetConfig.ToConfig(),
		StatefulSetConfig.ToConfig(),
		ConfigMapConfig.ToConfig(),
		PodConfig.ToConfig(),
		SecretConfig.ToConfig(),
		PersistentVolumeConfig.ToConfig(),
		PersistentVolumeClaimConfig.ToConfig(),
	}
	re := make([]processor.Config, 0, len(raw))
	for _, c := range raw {
		if c != nil {
			re = append(re, *c)
		}
	}
	return re
}
