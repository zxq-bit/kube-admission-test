package demo

import (
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	acorev1 "github.com/zxq-bit/kube-admission-test/pkg/admission/framework/interfaces/apis/core/v1"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ModelName = "demo"
)

func GetConfigMapStaticProcessor() *acorev1.ConfigMapProcessor {
	return &acorev1.ConfigMapProcessor{
		Metadata: processor.Metadata{
			Name:      "static",
			ModelName: ModelName,
			IgnoreNamespaces: []string{
				//metav1.NamespaceDefault,
				metav1.NamespaceSystem,
			},
			Type: constants.ProcessorTypeMutate,
		},
		Review: func(in *corev1.ConfigMap) (err error) {
			if in.Annotations == nil {
				in.Annotations = map[string]string{}
			}
			in.Annotations["mutated"] = "true"
			return
		},
	}
}
