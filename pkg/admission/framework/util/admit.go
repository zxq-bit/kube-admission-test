package util

import (
	"context"

	"github.com/caicloud/clientset/kubernetes/scheme"
	"github.com/caicloud/nirvana/log"

	admissionv1b1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AdmissionReviewFunc func(ctx context.Context, req *admissionv1b1.AdmissionReview) (resp *admissionv1b1.AdmissionReview)
type AdmitFunc func(ctx context.Context, obj metav1.Object) (err error)

func MakeAdmissionServeFunc(ctx context.Context, fs []AdmitFunc) AdmissionReviewFunc {
	return func(ctx context.Context, ar *admissionv1b1.AdmissionReview) (resp *admissionv1b1.AdmissionReview) {
		auid := ar.Request.UID

		deserializer := scheme.Codecs.UniversalDeserializer()
		if _, _, err := deserializer.Decode(data, nil, &ar); err != nil {
			log.Error(err)
			reviewResponse = toAdmissionResponse(err)
		} else {
			reviewResponse = s.admit(ar)
		}
		// 4. write result to response
		response := admissionv1b1.AdmissionReview{}
		if reviewResponse != nil {
			response.Response = reviewResponse
			response.Response.UID = ar.Request.UID
		}
	}
}
