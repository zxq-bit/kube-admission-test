package util

import (
	"fmt"
	"strings"

	"github.com/caicloud/go-common/slice"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// start options about

// filter by model name, return true if enabled
type ModelFilter func(modelName string) bool

func MakeModelFilter(modelNames []string) ModelFilter {
	return makeFilterFromValues(modelNames)
}

// ignore filter about

// ignore filter by namespace, return true if in ignore list
type NamespaceIgnoreFilter func(nsName string) bool

// ignore filter by annotations, return kv pair in ignore list
type AnnotationsIgnoreFilter func(anno map[string]string) (kv *[2]string)

// ignore filter of metav1 object, return a reason if should be ignore
type ObjectIgnoreFilter func(obj metav1.Object) (ignoreReason *string)

// ignore sets
type IgnoreSetting struct {
	// IgnoredNamespaces set namespaces that will be ignored
	Namespaces []string
	// AnnotationsValue set annotations value that objects who has it will be ignored
	AnnotationsValue string
	// TODO
	// OwnerReferences set owner setting that objects who match them will be ignored
	OwnerReferences []metav1.OwnerReference
}

func (is *IgnoreSetting) GetObjectFilter(annoKey string) ObjectIgnoreFilter {
	var filters []ObjectIgnoreFilter
	// namespace
	if len(is.Namespaces) > 0 {
		filters = append(filters, MakeNamespaceIgnoreObjectFilter(is.Namespaces))
	}
	// annotations
	if is.AnnotationsValue != "" && annoKey != "" {
		filters = append(filters, MakeAnnotationsIgnoreObjectFilter(annoKey, is.AnnotationsValue))
	}
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

func MakeNamespaceIgnoreFilter(nsNames []string) NamespaceIgnoreFilter {
	return makeFilterFromValues(nsNames)
}

func MakeAnnotationsIgnoreFilter(key, value string) AnnotationsIgnoreFilter {
	return func(anno map[string]string) *[2]string {
		if len(anno) == 0 {
			return nil
		}
		valueString, ok := anno[key]
		if !ok {
			return nil
		}
		values := splitAnnotationsFilterValues(valueString)
		annoValueFilter := makeFilterFromValues(values)
		if annoValueFilter(value) {
			return &[2]string{key, value}
		}
		return nil
	}
}

func MakeNamespaceIgnoreObjectFilter(nsNames []string) ObjectIgnoreFilter {
	nsFilter := MakeNamespaceIgnoreFilter(nsNames)
	return func(obj metav1.Object) (ignoreReason *string) {
		if nsName := obj.GetNamespace(); nsFilter(nsName) {
			reason := fmt.Sprintf("object namespace '%s' is in ingore list", nsName)
			ignoreReason = &reason
		}
		return
	}
}

func MakeAnnotationsIgnoreObjectFilter(key, value string) ObjectIgnoreFilter {
	annoFilter := MakeAnnotationsIgnoreFilter(key, value)
	return func(obj metav1.Object) (ignoreReason *string) {
		if kv := annoFilter(obj.GetAnnotations()); kv != nil {
			reason := fmt.Sprintf("object annotations [%s:%s] is in ingore list", kv[0], kv[1])
			ignoreReason = &reason
		}
		return
	}
}

func splitAnnotationsFilterValues(valueString string) (values []string) {
	vec := strings.Split(valueString, AnnoValueSplitKey)
	if !slice.StringInSlice(vec, AnnoValueFilterMatchAll) {
		return vec
	}
	values = append(values, AnnoValueFilterMatchAll)
	for _, value := range vec {
		if strings.HasPrefix(value, AnnoValueReversePrefix) {
			values = append(values, value)
		}
	}
	return
}

func makeFilterFromValues(values []string) func(string) bool {
	m := make(map[string]struct{}, len(values))
	for _, s := range values {
		m[s] = struct{}{}
	}
	matchAll := slice.StringInSlice(values, AnnoValueFilterMatchAll)
	return func(s string) bool {
		reverse := AnnoValueReversePrefix + s
		_, ok := m[reverse]
		if ok {
			return false
		}
		if matchAll {
			return true
		}
		_, ok = m[s]
		return ok
	}
}
