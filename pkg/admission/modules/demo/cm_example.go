package demo

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	rcorev1 "github.com/zxq-bit/kube-admission-test/pkg/admission/framework/review/apis/core/v1"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/review/processor"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"

	"github.com/caicloud/go-common/kubernetes/meta"
	"github.com/caicloud/nirvana/log"

	corev1 "k8s.io/api/core/v1"
)

var (
	cmProcessorExample = makeCmExampleProcessor()
)

func makeCmExampleProcessor() *rcorev1.ConfigMapProcessor {
	p := &rcorev1.ConfigMapProcessor{
		Metadata: processor.Metadata{
			Name:             ProcessorNameCmExample,
			ModuleName:       ModuleName,
			IgnoreNamespaces: []string{},
			Type:             constants.ProcessorTypeMutate,
		},
	}
	p.Review = func(ctx context.Context, in *corev1.ConfigMap) (err error) {
		logPrefix := fmt.Sprintf("%s[%s]", util.GetContextLogBase(ctx), p.LogPrefix())
		old, err := rcorev1.GetContextOldConfigMap(ctx)
		if err != nil {
			log.Errorf("%s GetContextOldConfigMap failed, %v", logPrefix, err)
			return err
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
