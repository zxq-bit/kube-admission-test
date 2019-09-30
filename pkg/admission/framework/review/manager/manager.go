package manager

import (
	"github.com/caicloud/nirvana/definition"
	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/errors"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/review"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"
)

type Manager struct {
	handlerMap map[schema.GroupVersionResource]map[arv1b1.OperationType]review.Handler
}

func (m *Manager) GetHandler(gvr schema.GroupVersionResource, opType arv1b1.OperationType) (review.Handler, error) {
	if m.handlerMap == nil {
		m.handlerMap = map[schema.GroupVersionResource]map[arv1b1.OperationType]review.Handler{}
	}
	// get exist
	if m.handlerMap[gvr] == nil {
		m.handlerMap[gvr] = map[arv1b1.OperationType]review.Handler{}
	}
	h := m.handlerMap[gvr][opType]
	if h != nil {
		return h, nil
	}
	// get maker
	maker := review.GetHandlerMaker(gvr)
	if maker == nil {
		return nil, errors.ErrNoHandlerMakerGVR
	}
	// make handler
	h, e := maker(opType)
	if e != nil {
		return nil, e
	}
	// register
	m.handlerMap[gvr][opType] = h
	return h, nil
}

func (m *Manager) GetDescriptors() (re []definition.Descriptor) {
	if m.handlerMap == nil {
		return
	}
	for _, sub := range m.handlerMap {
		if len(sub) == 0 {
			continue
		}
		for _, h := range sub {
			if h == nil || h.IsEmpty() {
				continue
			}
			re = append(re, h.ToNirvanaDescriptors())
		}
	}
	return re
}

func (m *Manager) GetMutatingWebHooks(svcNamespace, svcName string, svcCA []byte) (re []arv1b1.MutatingWebhookConfiguration) {
	if m.handlerMap == nil {
		return
	}
	for gvr, sub := range m.handlerMap {
		if len(sub) == 0 {
			continue
		}
		whConfig := arv1b1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: util.JoinObjectName(gvr.Group, gvr.Version, gvr.Resource),
			},
			Webhooks: []arv1b1.Webhook{},
		}
		for _, h := range sub {
			if h != nil && !h.IsEmpty() {
				whConfig.Webhooks = append(whConfig.Webhooks, h.ToWebHook(svcNamespace, svcName, svcCA))
			}
		}
		if len(whConfig.Webhooks) == 0 {
			continue
		}
		re = append(re, whConfig)
	}
	return re
}
