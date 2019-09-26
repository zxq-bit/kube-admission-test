package processor

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"
)

// ignore filter about

// filter for Metadata, return true if object can be use
type MetadataFilter func(metadata *Metadata) bool

// processor metadata, set name, type and ignore settings
type Metadata struct {
	// Name describe the name of this processor, like deployment-workload-name
	// effective at annotation's ignore option
	Name string
	// ModelName describe the model of this config, like app or resource
	// effective at admission server ignore option
	ModelName string
	// IgnoredNamespaces set namespaces that will be ignored
	IgnoreNamespaces []string
	// IgnoreOwnerReferences set owner setting that objects who match them will be ignored
	// TODO
	IgnoreOwnerReferences []metav1.OwnerReference
	// Type Validate or Mutate, decide weather to allow input object changes
	Type constants.ProcessorType
}

func (meta *Metadata) Validate() error {
	if meta.Name == "" {
		return fmt.Errorf("empty processor name")
	}
	if meta.Type != constants.ProcessorTypeValidate && meta.Type != constants.ProcessorTypeMutate {
		return fmt.Errorf("%v invalid processor type %v", meta.Name, meta.Type)
	}
	return nil
}

func (meta *Metadata) GetObjectFilter() util.ObjectIgnoreFilter {
	var filters []util.ObjectIgnoreFilter
	// namespace
	if len(meta.IgnoreNamespaces) > 0 {
		filters = append(filters, util.MakeNamespaceIgnoreObjectFilter(meta.IgnoreNamespaces))
	}
	// name
	if meta.Name != "" {
		filters = append(filters, util.MakeNameIgnoreObjectFilter(constants.AnnoKeyAdmissionIgnore, meta.Name))
	}
	// owner // TODO
	// return
	if len(filters) == 0 {
		return func(obj metav1.Object) *string {
			return nil
		}
	}
	return func(obj metav1.Object) *string {
		for _, f := range filters {
			if ignoreReason := f(obj); ignoreReason != nil {
				return ignoreReason
			}
		}
		return nil
	}
}

func MakeMetaFilterByModel(modelNames []string) MetadataFilter {
	modelFilter := util.MakeModelEnabledFilter(modelNames)
	return func(metadata *Metadata) bool {
		return modelFilter(metadata.ModelName)
	}
}
