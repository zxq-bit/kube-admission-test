package server

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/caicloud/go-common/cert"
	"github.com/caicloud/go-common/kubernetes/errors"
	"github.com/caicloud/nirvana/log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"
)

func (s *Server) ensureCert() (caBundle []byte, certFile, keyFile string, err error) {
	certData, keyData, err := cert.GenSelfSignedCertForK8sService(s.cfg.ServiceNamespace, s.cfg.ServiceName)
	if err != nil {
		return
	}
	caBundle = certData
	certFile = filepath.Join(s.cfg.CertTempDir, constants.DefaultCertFileName)
	keyFile = filepath.Join(s.cfg.CertTempDir, constants.DefaultKeyFileName)
	dataPath := map[string][]byte{
		certFile: certData,
		keyFile:  keyData,
	}
	for fp, data := range dataPath {
		if err = ioutil.WriteFile(fp, data, 0664); err != nil {
			return
		}
	}
	return
}

func (s *Server) ensureWebhooks(opt *processor.StartOptions) error {
	webhooks := s.configCollection.GetMutatingWebHooks(opt)
	for _, webhook := range webhooks {
		logPrefix := fmt.Sprintf("ensureWebhook[name:%s]", webhook.Name)
		prev, err := s.kc.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Get(webhook.Name, metav1.GetOptions{})
		switch {
		case err == nil:
			// found: update it
			prev.Webhooks = webhook.Webhooks
			_, err = s.kc.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Update(prev)
			if err != nil {
				log.Errorf("%s update failed, %v", logPrefix, err)
			} else {
				log.Infof("%s update done", logPrefix)
			}
		case errors.IsNotFound(err):
			// not found: create it
			_, err = s.kc.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Create(&webhook)
			if err != nil {
				log.Errorf("%s create failed, %v", logPrefix, err)
			} else {
				log.Infof("%s create done", logPrefix)
			}
		default:
			log.Errorf("%s get failed, %v", logPrefix, err)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
