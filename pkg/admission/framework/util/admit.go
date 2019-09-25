package util

import (
	"context"
	"encoding/json"

	"github.com/mattbaird/jsonpatch"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	admissionv1b1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ReviewFunc func(in runtime.Object) (err error)
type ReviewFuncWithContext func(ctx context.Context, in runtime.Object) (err error)
type AdmissionNirvanaFunc func(ctx context.Context, req *admissionv1b1.AdmissionReview) (resp *admissionv1b1.AdmissionReview)

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
