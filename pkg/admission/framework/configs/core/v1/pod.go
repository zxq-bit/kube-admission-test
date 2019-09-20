package v1

import (
	"fmt"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"
	"k8s.io/apimachinery/pkg/runtime"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Group    = ""
	Version  = "v1"
	Resource = "pods"
)

type PodProcessor struct {
	Name string
	// IgnoreSetting set namespaces and annotations that will ignore this processor
	IgnoreSetting *util.IgnoreSetting
	// Type Validate or Mutate, decide weather to allow input object changes
	Type util.ProcessorType
	// Review do review, return error if should stop
	Review func(in *corev1.Pod) (err error)
}

type Config struct {
	// ModelName describe the model of this config, like app or resource
	ModelName string
	// ProcessorsMap map pod processors by operation type
	ProcessorsMap map[arv1b1.OperationType][]PodProcessor
}

func (c *Config) Register(opType arv1b1.OperationType, p PodProcessor) {
	if c.ProcessorsMap == nil {
		c.ProcessorsMap = map[arv1b1.OperationType][]PodProcessor{}
	}
	c.ProcessorsMap[opType] = append(c.ProcessorsMap[opType], p)
}

func (c *Config) ToConfig() (out *util.Config) {
	out = &util.Config{
		ModelName: c.ModelName,
		GroupVersionResource: metav1.GroupVersionResource{
			Group:    Group,
			Version:  Version,
			Resource: Resource,
		},
		ProcessorsMap: make(map[arv1b1.OperationType][]util.Processor, len(c.ProcessorsMap)),
	}
	for opType, pps := range c.ProcessorsMap {
		if len(pps) == 0 {
			continue
		}
		out.ProcessorsMap[opType] = make([]util.Processor, 0, len(pps))
		for i := range pps {
			p := &pps[i]
			out.ProcessorsMap[opType] = append(out.ProcessorsMap[opType], util.Processor{
				Name:          p.Name,
				IgnoreSetting: p.IgnoreSetting,
				Type:          p.Type,
				Review: func(in runtime.Object) (err error) {
					pod := in.(*corev1.Pod)
					if pod == nil {
						err = fmt.Errorf("%s failed for input Pod is nil", p.Name)
					} else {
						err = p.Review(pod)
					}
					return err
				},
			})
		}
	}
	return out
}
