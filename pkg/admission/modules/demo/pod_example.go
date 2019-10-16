package demo

import (
	"context"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/errors"
	rcorev1 "github.com/zxq-bit/kube-admission-test/pkg/admission/framework/review/apis/core/v1"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/review/processor"

	corev1 "k8s.io/api/core/v1"
)

var (
	podProcessorExample = &rcorev1.PodProcessor{
		Metadata: processor.Metadata{
			Name:             ProcessorNamePodExample,
			ModuleName:       ModuleName,
			IgnoreNamespaces: []string{},
			Type:             constants.ProcessorTypeMutate,
		},
		Admit: func(ctx context.Context, in *corev1.Pod) (ke errors.APIStatus) {
			if in.Annotations == nil {
				in.Annotations = map[string]string{}
			}
			in.Annotations["mutated"] = "true"
			return
		},
	}
)
