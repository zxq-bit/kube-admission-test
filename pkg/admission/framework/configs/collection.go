package configs

import (
	cfgappsv1 "github.com/zxq-bit/kube-admission-test/pkg/admission/framework/configs/apis/apps/v1"
	cfgcorev1 "github.com/zxq-bit/kube-admission-test/pkg/admission/framework/configs/apis/core/v1"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"

	"github.com/caicloud/nirvana/definition"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
)

type Collection struct {
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

func (c *Collection) GetConfigs() []processor.Config {
	raw := []*processor.Config{
		c.DaemonSetConfig.ToConfig(),
		c.DeploymentConfig.ToConfig(),
		c.ReplicaSetConfig.ToConfig(),
		c.StatefulSetConfig.ToConfig(),
		c.ConfigMapConfig.ToConfig(),
		c.PodConfig.ToConfig(),
		c.SecretConfig.ToConfig(),
		c.PersistentVolumeConfig.ToConfig(),
		c.PersistentVolumeClaimConfig.ToConfig(),
	}
	re := make([]processor.Config, 0, len(raw))
	for _, c := range raw {
		if c != nil {
			re = append(re, *c)
		}
	}
	return re
}

func (c *Collection) GetDescriptors(opt *processor.StartOptions) (re []definition.Descriptor) {
	pcs := c.GetConfigs()
	for _, pc := range pcs {
		ds := pc.ToNirvanaDescriptors(opt)
		if len(ds) > 0 {
			re = append(re, ds...)
		}
	}
	return re
}

func (c *Collection) GetMutatingWebHooks(opt *processor.StartOptions) (re []arv1b1.MutatingWebhookConfiguration) {
	pcs := c.GetConfigs()
	for _, pc := range pcs {
		wc := pc.ToMutatingWebHook(opt)
		if wc != nil {
			re = append(re, *wc)
		}
	}
	return re
}
