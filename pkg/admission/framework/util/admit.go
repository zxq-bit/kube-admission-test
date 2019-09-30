package util

import (
	"encoding/json"

	"github.com/mattbaird/jsonpatch"
	admissionv1b1 "k8s.io/api/admission/v1beta1"
	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
)

func IsOperationTypeLeague(opType arv1b1.OperationType) bool {
	for i := range constants.OperationTypes {
		if opType == constants.OperationTypes[i] {
			return true
		}
	}
	return false
}

func ToAdmissionFailedResponse(uid types.UID, err error) *admissionv1b1.AdmissionResponse {
	switch e := err.(type) {
	case errors.APIStatus:
		status := e.Status()
		return &admissionv1b1.AdmissionResponse{
			UID:     uid,
			Allowed: false,
			Result:  &status,
		}
	default:
		return &admissionv1b1.AdmissionResponse{
			UID:     uid,
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}
}

func ToAdmissionPassResponse(uid types.UID, org, obj runtime.Object) *admissionv1b1.AdmissionResponse {
	orgJSON, err := json.Marshal(org)
	if err != nil {
		return ToAdmissionFailedResponse(uid, err)

	}
	objJSON, err := json.Marshal(obj)
	if err != nil {
		return ToAdmissionFailedResponse(uid, err)
	}

	rawPatch, err := jsonpatch.CreatePatch(orgJSON, objJSON)
	if err != nil {
		return ToAdmissionFailedResponse(uid, err)
	}

	patch, err := json.Marshal(rawPatch)
	if err != nil {
		return ToAdmissionFailedResponse(uid, err)
	}

	patchType := admissionv1b1.PatchTypeJSONPatch
	return &admissionv1b1.AdmissionResponse{
		UID:       uid,
		Allowed:   true,
		Patch:     patch,
		PatchType: &patchType,
	}
}
