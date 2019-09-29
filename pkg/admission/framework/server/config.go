package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"

	"github.com/caicloud/go-common/cert"
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

func (c *Config) String() string {
	b, _ := json.MarshalIndent(c, "", "  ")
	return string(b)
}

func (s *Server) validateConfig() (e error) {
	if s.cfg.InformerFactoryResync < 0 {
		return fmt.Errorf("informer factory resync setting %v is invalid", s.cfg.InformerFactoryResync)
	}
	if s.cfg.ServiceNamespace == "" {
		return fmt.Errorf("service namespace setting is empty")
	}
	if s.cfg.ServiceName == "" {
		return fmt.Errorf("service name setting is empty")
	}
	if s.cfg.CertTempDir == "" {
		return fmt.Errorf("cert temp dir setting is empty")
	}
	if s.enableOptions, e = s.parsedEnableOptions(); e != nil {
		return e
	}
	if s.serviceSelector, e = s.parsedServiceSelector(); e != nil {
		return e
	}
	if e = s.ensureCert(); e != nil {
		return e
	}
	return nil
}

func (s *Server) getStartOptions() processor.StartOptions {
	opt := processor.StartOptions{
		EnableOptions:    s.enableOptions,
		ServiceNamespace: s.cfg.ServiceNamespace,
		ServiceName:      s.cfg.ServiceName,
		ServiceCABundle:  s.caBundle,
		APIRootPath:      constants.DefaultAPIRootPath,
	}
	if len(opt.EnableOptions) == 0 {
		opt.EnableOptions, _ = s.parsedEnableOptions()
	}
	return opt
}

func (s *Server) ensureCert() error {
	certData, keyData, err := cert.GenSelfSignedCertForK8sService(s.cfg.ServiceNamespace, s.cfg.ServiceName)
	if err != nil {
		return err
	}
	s.caBundle = certData
	s.certFile = filepath.Join(s.cfg.CertTempDir, constants.DefaultCertFileName)
	s.keyFile = filepath.Join(s.cfg.CertTempDir, constants.DefaultKeyFileName)
	dataPath := map[string][]byte{
		s.certFile: certData,
		s.keyFile:  keyData,
	}
	for fp, data := range dataPath {
		if err = ioutil.WriteFile(fp, data, 0664); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) parsedEnableOptions() ([]string, error) {
	if s.cfg.Admissions == "" {
		return nil, fmt.Errorf("admissions setting is empty")
	}
	return strings.Split(s.cfg.Admissions, constants.AdmissionSplitKey), nil
}

func (s *Server) parsedServiceSelector() (map[string]string, error) {
	if s.cfg.ServiceSelector == "" {
		return nil, fmt.Errorf("service selector is empty")
	}
	kvs := strings.Split(s.cfg.ServiceSelector, constants.SelectorSplitKey)
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
