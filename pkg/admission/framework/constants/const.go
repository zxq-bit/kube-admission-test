package constants

import (
	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
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
