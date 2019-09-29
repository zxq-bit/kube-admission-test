package server

import (
	"fmt"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"

	"github.com/caicloud/go-common/kubernetes/errors"
	"github.com/caicloud/nirvana/log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

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

func (s *Server) ensureService(port int) error {
	next := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.cfg.ServiceNamespace,
			Name:      s.cfg.ServiceName,
		},
		Spec: corev1.ServiceSpec{
			Selector: s.serviceSelector,
			Ports: []corev1.ServicePort{
				{
					Name:       "https",
					Protocol:   corev1.ProtocolTCP,
					Port:       443,
					TargetPort: intstr.FromInt(port),
				},
			},
		},
	}
	logPrefix := fmt.Sprintf("ensureService[%s/%s:%d][%v]", next.Namespace, next.Name, port, next.Spec.Selector)
	prev, err := s.kc.CoreV1().Services(next.Namespace).Get(next.Name, metav1.GetOptions{})
	switch {
	case err == nil:
		// found: update it
		prev.Spec.Selector = next.Spec.Selector
		prev.Spec.Ports = next.Spec.Ports
		_, err = s.kc.CoreV1().Services(next.Namespace).Update(prev)
		if err != nil {
			log.Errorf("%s update failed, %v", logPrefix, err)
		} else {
			log.Infof("%s update done", logPrefix)
		}
	case errors.IsNotFound(err):
		// not found: create it
		_, err = s.kc.CoreV1().Services(next.Namespace).Create(next)
		if err != nil {
			log.Errorf("%s create failed, %v", logPrefix, err)
		} else {
			log.Infof("%s create done", logPrefix)
		}
	default:
		log.Errorf("%s get failed, %v", logPrefix, err)
	}
	return err
}
