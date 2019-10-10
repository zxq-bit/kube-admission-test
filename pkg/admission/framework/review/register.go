package review

import (
	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type HandlerTypeMaker func(operationType arv1b1.OperationType) (Handler, error)

var handlerRegister = map[schema.GroupVersionResource]HandlerTypeMaker{}

func RegisterHandlerMaker(gvr schema.GroupVersionResource, f HandlerTypeMaker) {
	handlerRegister[gvr] = f
}

func GetHandlerMaker(gvr schema.GroupVersionResource) HandlerTypeMaker {
	return handlerRegister[gvr]
}
