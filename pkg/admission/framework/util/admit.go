package util

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"

	"github.com/mattbaird/jsonpatch"
	admissionv1b1 "k8s.io/api/admission/v1beta1"
	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func IsOperationTypeLeague(opType arv1b1.OperationType) bool {
	for i := range constants.OperationTypes {
		if opType == constants.OperationTypes[i] {
			return true
		}
	}
	return false
}

func ToAdmissionFailedResponse(uid types.UID, kubeErr errors.APIStatus) *admissionv1b1.AdmissionResponse {
	status := kubeErr.Status()
	return &admissionv1b1.AdmissionResponse{
		UID:     uid,
		Allowed: false,
		Result:  &status,
	}
}

func ToAdmissionPassResponse(uid types.UID, org, obj runtime.Object) *admissionv1b1.AdmissionResponse {
	if org == nil || obj == nil {
		return &admissionv1b1.AdmissionResponse{
			UID:     uid,
			Allowed: true,
		}
	}
	orgJSON, err := json.Marshal(org)
	if err != nil {
		return ToAdmissionFailedResponse(uid, errors.NewBadRequest(err.Error()))

	}
	objJSON, err := json.Marshal(obj)
	if err != nil {
		return ToAdmissionFailedResponse(uid, errors.NewBadRequest(err.Error()))
	}

	rawPatch, err := jsonpatch.CreatePatch(orgJSON, objJSON)
	if err != nil {
		return ToAdmissionFailedResponse(uid, errors.NewInternalError(err))
	}

	patch, err := json.Marshal(rawPatch)
	if err != nil {
		return ToAdmissionFailedResponse(uid, errors.NewInternalError(err))
	}

	patchType := admissionv1b1.PatchTypeJSONPatch
	return &admissionv1b1.AdmissionResponse{
		UID:       uid,
		Allowed:   true,
		Patch:     patch,
		PatchType: &patchType,
	}
}

func AdmissionRequestLogBase(ar *admissionv1b1.AdmissionRequest) string {
	if ar == nil {
		return ""
	}
	return fmt.Sprintf("[%s][%v][%s]",
		path.Join(ar.Resource.Group, ar.Resource.Version, ar.Resource.Resource),
		ar.Operation,
		path.Join(ar.Namespace, ar.Name))
}
