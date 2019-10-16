package util

import (
	"fmt"
	"strings"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"

	"github.com/caicloud/go-common/slice"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// filter for enable or disable

// EnabledFilter return true if enabled
type EnabledFilter func(name string) bool

func makeEnabledFilterFromValues(values []string) func(string) bool {
	m := make(map[string]struct{}, len(values))
	for _, s := range values {
		m[s] = struct{}{}
	}
	matchAll := slice.StringInSlice(values, constants.AnnoValueFilterMatchAll)
	return func(s string) bool {
		// in pass list
		_, ok := m[s]
		if ok {
			return true
		}
		// in reverse list
		reverse := constants.AnnoValueReversePrefix + s
		_, ok = m[reverse]
		if ok {
			return false
		}
		// not in any list, match all
		return matchAll
	}
}

func simpleEnabledFilter(values []string, name string) bool {
	matchAll := false
	reverse := constants.AnnoValueReversePrefix + name
	for _, s := range values {
		// in pass list
		if s == name {
			return true
		}
		// in reverse list
		if s == reverse {
			return false
		}
		// match all
		matchAll = matchAll || s == constants.AnnoValueFilterMatchAll
	}
	// not in any list, match all
	return matchAll
}

// filter for module name

// filter by module name, return true if enabled
type ModuleEnabledFilter EnabledFilter

func MakeModuleEnabledFilter(moduleNames []string) ModuleEnabledFilter {
	return makeEnabledFilterFromValues(moduleNames)
}

// filter for namespaces

// ignore filter by namespace, return true if in ignore list
type NamespaceIgnoreFilter func(name string) bool

func makeNamespaceIgnoreFilter(nsNames []string) NamespaceIgnoreFilter {
	return makeEnabledFilterFromValues(nsNames)
}

// filter for name

// ignore filter by annotations, return true if name in object ignore option from annotations
type NameIgnoreFilter func(objAnno map[string]string) bool

// makeNameIgnoreFilter make module/name filter, filter out name is in annotation
func makeNameIgnoreFilter(key, pathValue string) NameIgnoreFilter {
	return func(anno map[string]string) bool {
		if len(anno) == 0 {
			return false
		}
		valueString, ok := anno[key]
		if !ok {
			return false
		}
		values := splitAnnotationsFilterValues(valueString)
		return simpleEnabledFilter(values, pathValue)
	}
}

func splitAnnotationsFilterValues(valueString string) (values []string) {
	vec := strings.Split(valueString, constants.AnnoValueSplitKey)
	if !slice.StringInSlice(vec, constants.AnnoValueFilterMatchAll) {
		return vec
	}
	values = append(values, constants.AnnoValueFilterMatchAll)
	for _, value := range vec {
		if strings.HasPrefix(value, constants.AnnoValueReversePrefix) {
			values = append(values, value)
		}
	}
	return
}

// filter input is metav1 object

// ignore filter of metav1 object, return a reason if should be ignore
type ObjectIgnoreFilter func(obj metav1.Object) (ignoreReason *string)

func MakeNamespaceIgnoreObjectFilter(nsNames []string) ObjectIgnoreFilter {
	nsFilter := makeNamespaceIgnoreFilter(nsNames)
	return func(obj metav1.Object) (ignoreReason *string) {
		if nsName := obj.GetNamespace(); nsFilter(nsName) {
			reason := fmt.Sprintf("object namespace '%s' is in ingore list", nsName)
			ignoreReason = &reason
		}
		return
	}
}

func MakeNameIgnoreObjectFilter(key, pathValue string) ObjectIgnoreFilter {
	annoFilter := makeNameIgnoreFilter(key, pathValue)
	return func(obj metav1.Object) (ignoreReason *string) {
		if annoFilter(obj.GetAnnotations()) {
			reason := fmt.Sprintf("object annotations's ignore settings contain kv [%s:%s]", key, pathValue)
			ignoreReason = &reason
		}
		return
	}
}
