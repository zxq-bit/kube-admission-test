package processor

import (
	"context"
	"fmt"

	"github.com/caicloud/go-common/interfaces"
	"github.com/caicloud/nirvana/log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"
)

type Processor struct {
	// Metadata, set name, type and ignore settings
	Metadata
	// Review do review, return error if should stop
	Review util.ReviewFunc
}

func (p *Processor) FilterObject(obj metav1.Object) *string {
	return p.Metadata.GetObjectFilter()(obj)
}

func CombineProcessors(ps []Processor) util.ReviewFuncWithContext {
	return func(ctx context.Context, in runtime.Object) (err error) {
		if interfaces.IsNil(in) {
			return fmt.Errorf("nil input")
		}
		obj := in.(metav1.Object)
		if interfaces.IsNil(obj) {
			return fmt.Errorf("not metav1 object")
		}
		defer util.RemoveObjectAnno(obj, constants.AnnoKeyAdmissionIgnore)
		// always return a not nil `out`, if out is nil, use in
		for _, p := range ps {
			// check ignore
			if ignoreReason := p.FilterObject(obj); ignoreReason != nil {
				log.Infof("%s skip for %s", p.Name, *ignoreReason)
				continue
			}
			// do review
			select {
			case <-ctx.Done():
				err = fmt.Errorf("processor chain not finished correctly, context ended")
			default:
				switch p.Type {
				case constants.ProcessorTypeValidate:
					err = p.Review(in.DeepCopyObject())
				case constants.ProcessorTypeMutate:
					err = p.Review(in)
				default:
					log.Errorf("%s skip for unknown processor type '%v'", p.Type)
				}
			}
			if err != nil {
				break
			}
		}
		return
	}
}
