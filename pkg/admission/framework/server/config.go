package server

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"
)

type Config struct {
	// kube
	KubeHost   string `desc:"kubernetes host"`
	KubeConfig string `desc:"kubernetes config"`
	// informer
	InformerFactoryResync time.Duration `desc:"kubernetes informer factory resync time"`
	// admit
	ServiceNamespace string `desc:"admission service namespace"`
	ServiceName      string `desc:"admission service name"`
	ServiceSelector  string `desc:"admission service selector labels key value pairs"`
	CertTempDir      string `desc:"admission server cert file template dir path"`
	// enable
	Admissions string `desc:"a list of admissions to enable. '*' enables all on-by-default admissions"`

	enableOptions   []string
	serviceSelector map[string]string
}

func NewDefaultConfig() *Config {
	return &Config{
		InformerFactoryResync: constants.DefaultInformerFactoryResync,
		ServiceNamespace:      constants.DefaultServiceNamespace,
		ServiceName:           constants.DefaultServiceName,
		ServiceSelector:       constants.DefaultServiceSelector,
		CertTempDir:           constants.DefaultCertTempDir,
		Admissions:            constants.AdmissionsAll,
	}
}

func (c *Config) Validate() (e error) {
	if c.InformerFactoryResync < 0 {
		return fmt.Errorf("informer factory resync setting %v is invalid", c.InformerFactoryResync)
	}
	if c.ServiceNamespace == "" {
		return fmt.Errorf("service namespace setting is empty")
	}
	if c.ServiceName == "" {
		return fmt.Errorf("service name setting is empty")
	}
	if c.CertTempDir == "" {
		return fmt.Errorf("cert temp dir setting is empty")
	}
	if c.enableOptions, e = c.parsedEnableOptions(); e != nil {
		return e
	}
	if c.serviceSelector, e = c.parsedServiceSelector(); e != nil {
		return e
	}
	return nil
}

func (c *Config) String() string {
	b, _ := json.MarshalIndent(c, "", "  ")
	return string(b)
}

func (c *Config) ToStartOptions() processor.StartOptions {
	opt := processor.StartOptions{
		EnableOptions:    c.enableOptions,
		ServiceNamespace: c.ServiceNamespace,
		ServiceName:      c.ServiceName,
		APIRootPath:      constants.DefaultAPIRootPath,
	}
	if len(opt.EnableOptions) == 0 {
		opt.EnableOptions, _ = c.parsedEnableOptions()
	}
	return opt
}

func (c *Config) parsedEnableOptions() ([]string, error) {
	if c.Admissions == "" {
		return nil, fmt.Errorf("admissions setting is empty")
	}
	return strings.Split(c.Admissions, constants.AdmissionSplitKey), nil
}

func (c *Config) parsedServiceSelector() (map[string]string, error) {
	if c.ServiceSelector == "" {
		return nil, fmt.Errorf("service selector is empty")
	}
	kvs := strings.Split(c.ServiceSelector, constants.SelectorSplitKey)
	re := make(map[string]string, len(kvs))
	for i, s := range kvs {
		kv := strings.Split(s, constants.SelectorKVSplitKey)
		if len(kv) != 2 {
			return re, fmt.Errorf("kv pair %d in bad format: '%s'", i, s)
		}
		re[kv[0]] = kv[1]
	}
	return re, nil
}
