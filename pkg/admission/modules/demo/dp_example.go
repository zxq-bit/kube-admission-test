package demo

import (
	"context"
	"fmt"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/errors"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/review/gen"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/review/processor"

	appsv1 "k8s.io/api/apps/v1"
)

func (m *Module) getDpProcessorCheckMntRef() *gen.DeploymentProcessor {
	return &gen.DeploymentProcessor{
		Metadata: processor.Metadata{
			Name:             ProcessorNameDpCheckMntRef,
			ModuleName:       ModuleName,
			IgnoreNamespaces: []string{},
			Type:             constants.ProcessorTypeValidate,
		},
		Admit: m.dpReviewMntRef,
	}
}

func (m *Module) dpReviewMntRef(ctx context.Context, in *appsv1.Deployment) (ke *errors.StatusError) {
	for _, v := range in.Spec.Template.Spec.Volumes {
		if v.ConfigMap != nil {
			cmName := v.ConfigMap.Name
			_, e := m.configMapLister.ConfigMaps(in.Namespace).Get(cmName)
			if e != nil {
				return errors.NewBadRequest(fmt.Errorf("get dp ref cm %s failed, %v", cmName, e))
			}
		}
		if v.Secret != nil {
			secretName := v.Secret.SecretName
			_, e := m.configMapLister.ConfigMaps(in.Namespace).Get(secretName)
			if e != nil {
				return errors.NewBadRequest(fmt.Errorf("get dp ref secret %s failed, %v", secretName, e))
			}
		}
	}
	return nil
}
