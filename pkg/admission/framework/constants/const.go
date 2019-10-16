package constants

import (
	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultListenPort                  = 6666
	DefaultCertSecretNamespace         = metav1.NamespaceSystem
	DefaultCertSecretName              = "admission-cert"
	DefaultCertTempDir                 = "/tmp"
	DefaultCertFileName                = "admission.cert"
	DefaultKeyFileName                 = "admission.key"
	AdmissionsAll                      = "*"
	DefaultServiceNamespace            = metav1.NamespaceSystem
	DefaultServiceName                 = "admission-server-collection"
	DefaultInformerFactoryResyncSecond = 60
	DefaultReviewConfigFilePath        = "/caicloud/compass/admission/handler.yaml"

	AdmissionSplitKey = ","

	CertSecretKeyKey  = "key"
	CertSecretKeyCert = "cert"

	APIRootPath = "/apis/v1alpha1"
)

const (
	AnnoKeyAdmissionIgnore = "admission.caicloud.io/ignore"

	AnnoValueFilterMatchAll = "*"
	AnnoValueSplitKey       = ","
	AnnoValueReversePrefix  = "-"

	ProcessorKeySplit = "/"
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

type ContextKey string

const (
	ContextKeyLogBase          ContextKey = "logBase"
	ContextKeyAdmissionRequest ContextKey = "admissionRequest"
	ContextKeyObjectBackup     ContextKey = "objectBackup"
)
