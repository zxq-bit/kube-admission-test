package server

import (
	"fmt"

	"github.com/caicloud/go-common/kubernetes/errors"
	"github.com/caicloud/nirvana/log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Server) ensureWebhooks() error {
	webhooks := s.reviewManager.GetMutatingWebHooks(s.cfg.ServiceNamespace, s.cfg.ServiceName, s.caBundle)
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
