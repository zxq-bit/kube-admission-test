package constants

import (
	"time"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultListenPort            = 6666
	DefaultCertTempDir           = "/tmp"
	DefaultCertFileName          = "admission.cert"
	DefaultKeyFileName           = "admission.key"
	AdmissionsAll                = "*"
	DefaultServiceNamespace      = metav1.NamespaceSystem
	DefaultServiceName           = "admission-server-collection"
	DefaultServiceSelector       = "zxq-app:admission-test,selector-test:test"
	DefaultInformerFactoryResync = 60 * time.Second

	SelectorSplitKey   = ","
	SelectorKVSplitKey = ":"
	AdmissionSplitKey  = ","

	DefaultAPIRootPath = "/apis/v1alpha1"
)

const (
	AnnoKeyAdmissionIgnore = "admission.caicloud.io/ignore"

	AnnoValueFilterMatchAll = "*"
	AnnoValueSplitKey       = ","
	AnnoValueReversePrefix  = "-"
)

var (
	OperationTypes = []arv1b1.OperationType{
		arv1b1.Create,
		arv1b1.Update,
		arv1b1.Delete,
		// arv1b1.OperationAll,
		// arv1b1.Connect,
	}
)

type ProcessorType string

const (
	ProcessorTypeValidate ProcessorType = "Validate"
	ProcessorTypeMutate   ProcessorType = "Mutate"
)
