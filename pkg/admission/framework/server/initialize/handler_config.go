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
	Module string
	Name   string
}

type Handler struct {
	schema.GroupVersionResource `yaml:",inline" json:",inline"`

	OpType        arv1b1.OperationType
	TimeoutSecond int32
	Processors    []Processor
}

type HandlerConfig struct {
	Handlers []Handler
}

func ReadHandlerConfigFromFile(fp string) (*HandlerConfig, error) {
	b, e := ioutil.ReadFile(fp)
	if e != nil {
		return nil, e
	}
	re := &HandlerConfig{}
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

func (c *HandlerConfig) String() string {
	b, _ := yaml.Marshal(c)
	return string(b)
}

func (c *HandlerConfig) Validate() error {
	if len(c.Handlers) == 0 {
		return fmt.Errorf("review processors is empty")
	}
	for i, r := range c.Handlers {
		if e := r.Validate(); e != nil {
			return fmt.Errorf("review [%d][%s] validate failed, %v", i, r.String(), e)
		}
	}
	return nil
}

func (p *Processor) String() string {
	return strings.Join([]string{p.Module, p.Name}, "/")
}

func (p *Processor) Validate() error {
	if p.Module == "" {
		return fmt.Errorf("processor module is empty")
	}
	if p.Name == "" {
		return fmt.Errorf("processor name is empty")
	}
	return nil
}

func (h *Handler) String() string {
	return strings.Join([]string{h.Group, h.Version, h.Resource, string(h.OpType)}, "/")
}

func (h *Handler) Validate() error {
	if h.GroupVersionResource.Empty() {
		return errors.ErrEmptyGVR
	}
	if !util.IsOperationTypeLeague(h.OpType) {
		return fmt.Errorf("review op type is illegal")
	}
	if h.TimeoutSecond < 0 || h.TimeoutSecond > 30 {
		return errors.ErrBadTimeoutSecond
	}
	if len(h.Processors) == 0 {
		return fmt.Errorf("review processors is empty")
	}
	for i, p := range h.Processors {
		if e := p.Validate(); e != nil {
			return fmt.Errorf("review processor [%d][%s] validate failed, %v", i, p.String(), e)
		}
	}
	return nil
}
