package v1

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"

	"github.com/caicloud/nirvana/log"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	deploymentGRV = appsv1.SchemeGroupVersion.WithResource("deployment")
	deploymentGVK = appsv1.SchemeGroupVersion.WithKind("Deployment")
)

type DeploymentProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Review do review, return error if should stop
	Review func(in *appsv1.Deployment) (err error)
}

type DeploymentConfig struct {
	// TimeoutMap set total execute time of processors
	TimeoutMap map[arv1b1.OperationType]time.Duration
	// ProcessorsMap map pod processors by operation type
	ProcessorsMap map[arv1b1.OperationType][]DeploymentProcessor
}

func (p *DeploymentProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Review == nil {
		return fmt.Errorf("%v nil processor review function", p.Name)
	}
	return nil
}

func (c *DeploymentConfig) Register(opType arv1b1.OperationType, ps ...*DeploymentProcessor) {
	if c.ProcessorsMap == nil {
		c.ProcessorsMap = make(map[arv1b1.OperationType][]DeploymentProcessor, 1)
	}
	if len(c.ProcessorsMap[opType]) == 0 {
		c.ProcessorsMap[opType] = make([]DeploymentProcessor, 0, len(ps))
	}
	for i, p := range ps {
		if p == nil {
			continue
		}
		if e := p.Validate(); e != nil {
			log.Errorf("appsv1.Deployment processor register failed for [%d.%s], %v", i, p.Name, e)
		}
		c.ProcessorsMap[opType] = append(c.ProcessorsMap[opType], *p)
	}
	return
}

func (c *DeploymentConfig) SetTimeout(opType arv1b1.OperationType, timeout time.Duration) {
	if c.TimeoutMap == nil {
		c.TimeoutMap = make(map[arv1b1.OperationType]time.Duration, 1)
	}
	c.TimeoutMap[opType] = timeout
}

func (c *DeploymentConfig) ToConfig() (out *processor.Config) {
	out = &processor.Config{
		GroupVersionResource: deploymentGRV,
		RawExtensionParser: func(raw *runtime.RawExtension) (runtime.Object, error) {
			obj, e := GetDeploymentFromRawExtension(raw)
			if e != nil {
				return nil, e
			}
			return obj, nil
		},
		TimeoutMap:    c.TimeoutMap,
		ProcessorsMap: make(map[arv1b1.OperationType][]processor.Processor, len(c.ProcessorsMap)),
	}
	if out.TimeoutMap == nil {
		out.TimeoutMap = map[arv1b1.OperationType]time.Duration{}
	}
	for opType, ps := range c.ProcessorsMap {
		if len(ps) == 0 {
			continue
		}
		ops := make([]processor.Processor, 0, len(ps))
		for i := range ps {
			p := &ps[i]
			ops = append(ops, processor.Processor{
				Metadata: p.Metadata,
				Review: func(obj runtime.Object) (err error) {
					in := obj.(*appsv1.Deployment)
					if in == nil {
						err = fmt.Errorf("%s failed for input is nil", p.Name)
					} else {
						err = p.Review(in)
					}
					return err
				},
			})
		}
		if len(ops) > 0 {
			out.ProcessorsMap[opType] = ops
		}
	}
	if len(out.ProcessorsMap) == 0 {
		return nil
	}
	return out
}

func GetDeploymentFromRawExtension(raw *runtime.RawExtension) (*appsv1.Deployment, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != deploymentGVK {
		return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), deploymentGVK.String())
	}
	if obj := raw.Object.(*appsv1.Deployment); obj != nil {
		return obj, nil
	}
	parsed := &appsv1.Deployment{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}
