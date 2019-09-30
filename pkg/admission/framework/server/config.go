package server

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/caicloud/go-common/cert"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/server/initialize"
)

func (s *Server) validateConfig() (e error) {
	if s.cfg.InformerFactoryResyncSecond < 0 {
		return fmt.Errorf("informer factory resync setting %v is invalid", s.cfg.InformerFactoryResyncSecond)
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
	if s.reviewConfig, e = initialize.ReadReviewConfigFromFile(s.cfg.ReviewConfigFile); e != nil {
		return e
	}
	return nil
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
