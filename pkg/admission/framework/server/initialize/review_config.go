package initialize

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/errors"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"

	"github.com/caicloud/nirvana/log"

	"gopkg.in/yaml.v2"
	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Processor struct {
	Model string
	Name  string
}

type Review struct {
	schema.GroupVersionResource `yaml:",inline" json:",inline"`

	OpType        arv1b1.OperationType
	TimeoutSecond int32
	Processors    []Processor
}

type ReviewConfig struct {
	Reviews []Review
}

func ReadReviewConfigFromFile(fp string) (*ReviewConfig, error) {
	b, e := ioutil.ReadFile(fp)
	if e != nil {
		return nil, e
	}
	re := &ReviewConfig{}
	if filepath.Ext(fp) == "json" {
		e = json.Unmarshal(b, re)
	} else {
		e = yaml.Unmarshal(b, re)
	}
	if e != nil {
		return nil, e
	}
	log.Infof("parsed review config:\n%s", re.String())
	if e = re.Validate(); e != nil {
		return nil, e
	}
	return re, nil
}

func (c *ReviewConfig) String() string {
	b, _ := yaml.Marshal(c)
	return string(b)
}

func (c *ReviewConfig) Validate() error {
	if len(c.Reviews) == 0 {
		return fmt.Errorf("review processors is empty")
	}
	for i, r := range c.Reviews {
		if e := r.Validate(); e != nil {
			return fmt.Errorf("review [%d][%s] validate failed, %v", i, r.String(), e)
		}
	}
	return nil
}

func (p *Processor) String() string {
	return strings.Join([]string{p.Model, p.Name}, "/")
}

func (p *Processor) Validate() error {
	if p.Model == "" {
		return fmt.Errorf("processor model is empty")
	}
	if p.Name == "" {
		return fmt.Errorf("processor name is empty")
	}
	return nil
}

func (r *Review) String() string {
	return strings.Join([]string{r.Group, r.Version, r.Resource, string(r.OpType)}, "/")
}

func (r *Review) Validate() error {
	if r.GroupVersionResource.Empty() {
		return errors.ErrEmptyGVR
	}
	if !util.IsOperationTypeLeague(r.OpType) {
		return fmt.Errorf("review op type is illegal")
	}
	if r.TimeoutSecond < 0 || r.TimeoutSecond > 30 {
		return errors.ErrBadTimeoutSecond
	}
	if len(r.Processors) == 0 {
		return fmt.Errorf("review processors is empty")
	}
	for i, p := range r.Processors {
		if e := p.Validate(); e != nil {
			return fmt.Errorf("review processor [%d][%s] validate failed, %v", i, p.String(), e)
		}
	}
	return nil
}
