package server

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/server/initialize"

	"github.com/caicloud/go-common/cert"
	"github.com/caicloud/go-common/kubernetes/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	if s.cfg.CertSecretNamespace == "" {
		return fmt.Errorf("cert secret namespace is empty")
	}
	if s.cfg.CertSecretName == "" {
		return fmt.Errorf("cert secret name is empty")
	}
	if s.enableOptions, e = s.parsedEnableOptions(); e != nil {
		return e
	}
	if e = s.ensureCert(); e != nil {
		return e
	}
	if s.handlerConfig, e = initialize.ReadHandlerConfigFromFile(s.cfg.ReviewConfigFile); e != nil {
		return e
	}
	return nil
}

func (s *Server) ensureCert() error {
	// ensure secret
	secret, err := s.ensureCertSecret()
	for errors.IsNotFound(err) || errors.IsAlreadyExists(err) || errors.IsConflict(err) {
		secret, err = s.ensureCertSecret()
	}
	if err != nil {
		return err
	}
	certData, keyData := secret.Data[constants.CertSecretKeyCert], secret.Data[constants.CertSecretKeyKey]
	// ensure cert files
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

func (s *Server) ensureCertSecret() (*corev1.Secret, error) {
	ns, name := s.cfg.CertSecretNamespace, s.cfg.CertSecretName
	prev, err := s.kc.CoreV1().Secrets(ns).Get(name, metav1.GetOptions{})
	// prev is ok
	if validateCertSecret(prev) == nil {
		return prev, nil
	}
	// other error
	prevNotFound := errors.IsNotFound(err)
	if !prevNotFound {
		return nil, err
	}
	// gen
	certData, keyData, err := cert.GenSelfSignedCertForK8sService(s.cfg.ServiceNamespace, s.cfg.ServiceName)
	if err != nil {
		return nil, err
	}
	newData := map[string][]byte{
		constants.CertSecretKeyCert: certData,
		constants.CertSecretKeyKey:  keyData,
	}
	// need create
	if prevNotFound {
		return s.kc.CoreV1().Secrets(ns).Create(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      name,
			},
			Data: newData,
		})
	}
	// need update
	prev.Data = newData
	return s.kc.CoreV1().Secrets(ns).Update(prev)
}

func validateCertSecret(in *corev1.Secret) error {
	if in == nil {
		return fmt.Errorf("nil input secret")
	}
	if len(in.Data) == 0 {
		return fmt.Errorf("empty secret data")
	}
	for _, k := range []string{constants.CertSecretKeyKey, constants.CertSecretKeyCert} {
		if len(in.Data[k]) == 0 {
			return fmt.Errorf("empty '%s' in secret data", k)
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
