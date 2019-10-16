package demo

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/errors"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/review/gen"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/review/processor"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"

	"github.com/caicloud/go-common/kubernetes/meta"
	"github.com/caicloud/nirvana/log"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

var (
	cmProcessorExample       = makeCmExampleProcessor()
	cmProcessorDeletionAllow = makeCmDeletionAllowProcessor()
)

func makeCmExampleProcessor() *gen.ConfigMapProcessor {
	p := &gen.ConfigMapProcessor{
		Metadata: processor.Metadata{
			Name:             ProcessorNameCmExample,
			ModuleName:       ModuleName,
			IgnoreNamespaces: []string{},
			Type:             constants.ProcessorTypeMutate,
		},
	}
	p.Admit = func(ctx context.Context, in *corev1.ConfigMap) (ke errors.APIStatus) {
		logPrefix := fmt.Sprintf("%s[%s]", util.GetContextLogBase(ctx), p.LogPrefix())
		old, err := gen.GetContextOldConfigMap(ctx)
		if err != nil {
			log.Errorf("%s GetContextOldConfigMap failed, %v", logPrefix, err)
			return errors.NewBadRequest(err)
		}
		const (
			annoKey = "demo.caicloud.io/process-version"
		)
		// prev value
		prevVer := -1
		if parsed, e := strconv.Atoi(meta.GetAnnotation(old, annoKey)); e == nil && parsed >= 0 {
			prevVer = parsed
			log.Infof("%s parse prev %d", logPrefix, prevVer)
		}
		// new value
		if in.Annotations == nil {
			in.Annotations = map[string]string{}
		}
		nextVer := prevVer + 1
		in.Annotations[annoKey] = strconv.Itoa(nextVer)
		log.Infof("%s update to %d", logPrefix, nextVer)
		return
	}
	return p
}

func makeCmDeletionAllowProcessor() *gen.ConfigMapProcessor {
	p := &gen.ConfigMapProcessor{
		Metadata: processor.Metadata{
			Name:             ProcessorNameCmDeletionAllow,
			ModuleName:       ModuleName,
			IgnoreNamespaces: []string{},
			Type:             constants.ProcessorTypeValidate,
		},
	}
	p.Admit = func(ctx context.Context, in *corev1.ConfigMap) (ke errors.APIStatus) {
		logPrefix := fmt.Sprintf("%s[%s]", util.GetContextLogBase(ctx), p.LogPrefix())
		if opType := util.GetContextOpType(ctx); opType != arv1b1.Delete {
			log.Errorf("%s got unexpected op type: '%v'", logPrefix, opType)
			return errors.NewBadRequest(fmt.Errorf("unexpected op type '%v'", opType))
		}
		old, err := gen.GetContextOldConfigMap(ctx)
		if err != nil {
			log.Errorf("%s GetContextOldConfigMap failed, %v", logPrefix, err)
			return errors.NewInternalServerError(err)
		}

		const (
			annoKey = "demo.caicloud.io/deletion-not-allowed"
		)
		log.Infof("%s old: %s", logPrefix, toJson(old))
		log.Infof("%s new: %s", logPrefix, toJson(in))
		toCheck := old
		if in != nil {
			toCheck = in
		}
		notAllowed, _ := strconv.ParseBool(meta.GetAnnotation(toCheck, annoKey))
		if notAllowed {
			log.Errorf("%s not allowed to del", logPrefix)
			return errors.NewBadRequest(fmt.Errorf("deletion not allowed"))
		}
		log.Infof("%s allowed to del", logPrefix)
		return nil
	}
	return p
}

func toJson(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
