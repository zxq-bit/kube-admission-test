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
	// execute timeout by second
	TimeoutSecond int32
	// Processors in seq
	Processors []Processor
}

type GroupVersionResource struct {
	// GVR
	schema.GroupVersionResource `yaml:",inline" json:",inline"`
	// Handlers by operation type
	Handlers map[arv1b1.OperationType]*Handler
}

type HandlerConfig struct {
	Configs []GroupVersionResource
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
	if len(c.Configs) == 0 {
		return fmt.Errorf("review processors is empty")
	}
	for i, s := range c.Configs {
		if e := s.Validate(); e != nil {
			return fmt.Errorf("review [%d][%s] validate failed, %v", i, s.String(), e)
		}
	}
	return nil
}

func (s *GroupVersionResource) String() string {
	return strings.Join([]string{s.Group, s.Version, s.Resource}, "/")
}

func (s *GroupVersionResource) Validate() error {
	if s.GroupVersionResource.Empty() {
		return errors.ErrEmptyGVR
	}
	if len(s.Handlers) == 0 {
		return fmt.Errorf("no handler in set")
	}
	for opType, h := range s.Handlers {
		if !util.IsOperationTypeLeague(opType) {
			return fmt.Errorf("op type %v is illegal", opType)
		}
		if e := h.Validate(); e != nil {
			return fmt.Errorf("handler [%s] validate failed, %v", opType, e)
		}
	}
	return nil
}

func (h *Handler) Validate() error {
	if h.TimeoutSecond < 0 || h.TimeoutSecond > 30 {
		return errors.ErrBadTimeoutSecond
	}
	if len(h.Processors) == 0 {
		return fmt.Errorf("handler processors is empty")
	}
	for i, p := range h.Processors {
		if e := p.Validate(); e != nil {
			return fmt.Errorf("handler processor [%d][%s] validate failed, %v", i, p.String(), e)
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
