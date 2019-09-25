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
	CertTempDir      string `desc:"admission server cert file template dir path"`
	// enable
	Admissions string `desc:"a list of admissions to enable. '*' enables all on-by-default admissions"`
}

func NewDefaultConfig() *Config {
	return &Config{
		InformerFactoryResync: constants.DefaultInformerFactoryResync,
		ServiceNamespace:      constants.DefaultServiceNamespace,
		ServiceName:           constants.DefaultServiceName,
		CertTempDir:           constants.DefaultCertTempDir,
		Admissions:            constants.AdmissionsAll,
	}
}

func (c *Config) Validate() error {
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
	if c.Admissions == "" {
		return fmt.Errorf("admissions setting is empty")
	}
	return nil
}

func (c *Config) String() string {
	b, _ := json.MarshalIndent(c, "", "  ")
	return string(b)
}

func (c *Config) ToStartOptions() processor.StartOptions {
	return processor.StartOptions{
		EnableOptions:    strings.Split(c.Admissions, constants.AdmissionSplitKey),
		ServiceNamespace: c.ServiceNamespace,
		ServiceName:      c.ServiceName,
		APIRootPath:      constants.DefaultAPIRootPath,
	}
}
