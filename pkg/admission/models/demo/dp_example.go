package demo

import (
	"context"
	"fmt"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	rappsv1 "github.com/zxq-bit/kube-admission-test/pkg/admission/framework/review/apis/apps/v1"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/review/processor"

	appsv1 "k8s.io/api/apps/v1"
)

func (m *Model) getDpProcessorCheckMntRef() *rappsv1.DeploymentProcessor {
	return &rappsv1.DeploymentProcessor{
		Metadata: processor.Metadata{
			Name:             ProcessorNamePodExample,
			ModelName:        ModelName,
			IgnoreNamespaces: []string{},
			Type:             constants.ProcessorTypeValidate,
		},
		Review: m.dpReviewMntRef,
	}
}

func (m *Model) dpReviewMntRef(ctx context.Context, in *appsv1.Deployment) (err error) {
	for _, v := range in.Spec.Template.Spec.Volumes {
		if v.ConfigMap != nil {
			cmName := v.ConfigMap.Name
			_, e := m.configMapLister.ConfigMaps(in.Namespace).Get(cmName)
			if e != nil {
				return fmt.Errorf("get dp ref cm %s failed, %v", cmName, e)
			}
		}
		if v.Secret != nil {
			secretName := v.Secret.SecretName
			_, e := m.configMapLister.ConfigMaps(in.Namespace).Get(secretName)
			if e != nil {
				return fmt.Errorf("get dp ref secret %s failed, %v", secretName, e)
			}
		}
	}
	return nil
}