package processor

import (
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"
)

// start config about

type StartOptions struct {
	EnableOptions []string

	ServiceNamespace string
	ServiceName      string
	ServiceCABundle  []byte

	APIRootPath string
}

func (opt *StartOptions) GetProcessorFilter() (filter MetadataFilter) {
	modelFilter := util.MakeModelEnabledFilter(opt.EnableOptions)
	return func(metadata *Metadata) bool {
		return modelFilter(metadata.ModelName)
	}
}
