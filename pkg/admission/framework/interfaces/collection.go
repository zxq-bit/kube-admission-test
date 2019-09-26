package interfaces

import (
	"fmt"
	"sort"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"

	"github.com/caicloud/go-common/interfaces"
	"github.com/caicloud/nirvana/definition"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
)

func (c *ConfigCollection) GetConfigs(opt *processor.StartOptions) []processor.Config {
	processorFilter := opt.GetProcessorFilter()
	raw := []*processor.Config{
		c.DaemonSetConfig.ToConfig(processorFilter),
		c.DeploymentConfig.ToConfig(processorFilter),
		c.ReplicaSetConfig.ToConfig(processorFilter),
		c.StatefulSetConfig.ToConfig(processorFilter),
		c.ConfigMapConfig.ToConfig(processorFilter),
		c.PodConfig.ToConfig(processorFilter),
		c.SecretConfig.ToConfig(processorFilter),
		c.PersistentVolumeConfig.ToConfig(processorFilter),
		c.PersistentVolumeClaimConfig.ToConfig(processorFilter),
	}
	re := make([]processor.Config, 0, len(raw))
	for _, c := range raw {
		if c != nil {
			re = append(re, *c)
		}
	}
	return re
}

func (c *ConfigCollection) GetDescriptors(opt *processor.StartOptions) (re []definition.Descriptor) {
	pcs := c.GetConfigs(opt)
	for _, pc := range pcs {
		ds := pc.ToNirvanaDescriptors(opt)
		if len(ds) > 0 {
			re = append(re, ds...)
		}
	}
	return re
}

func (c *ConfigCollection) GetMutatingWebHooks(opt *processor.StartOptions) (re []arv1b1.MutatingWebhookConfiguration) {
	pcs := c.GetConfigs(opt)
	for _, pc := range pcs {
		wc := pc.ToMutatingWebHook(opt)
		if wc != nil {
			re = append(re, *wc)
		}
	}
	return re
}

func (c *ModelCollection) Register(models ...Model) error {
	for _, m := range models {
		if interfaces.IsNil(m) {
			return fmt.Errorf("model is nil")
		}
		if c.ModelMap == nil {
			c.ModelMap = make(map[string]Model, 1)
		}
		if _, exist := c.ModelMap[m.Name()]; exist {
			return fmt.Errorf("model %s already exist", m.Name())
		}
		c.ModelMap[m.Name()] = m
	}
	return nil
}

func (c *ModelCollection) ListModels() []Model {
	l := len(c.ModelMap)
	if l == 0 {
		return nil
	}
	re := make([]Model, 0, l)
	for _, m := range c.ModelMap {
		re = append(re, m)
	}
	sort.Slice(re, func(i, j int) bool {
		return re[i].Name() < re[j].Name()
	})
	return re
}
