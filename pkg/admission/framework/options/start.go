package options

import (
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/processor"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"
)

// start config about

type StartOptions struct {
	EnableOptions []string

	ServiceName      string
	ServiceNamespace string
	ServiceCABundle  []byte

	APIRootPath string
}

func (opt *StartOptions) FilterProcessorsByModel(in []processor.Processor) (out []processor.Processor) {
	isModelEnabled := util.MakeModelEnabledFilter(opt.EnableOptions)
	for _, p := range in {
		if isModelEnabled(p.ModelName) {
			out = append(out, p)
		}
	}
	return out
}
