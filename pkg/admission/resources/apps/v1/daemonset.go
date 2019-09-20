package resources

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/mattbaird/jsonpatch"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"github.com/caicloud/clientset/kubernetes/scheme"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/resources"
)

const (
	containerNetworkconfiguration = "v1.multus-cni.io/default-network"
)

func init() {
	r := metav1.GroupVersionResource{
		Group:    appsv1.GroupName,
		Version:  appsv1.SchemeGroupVersion.Version,
		Resource: "daemonsets",
	}
	d := &daemonsetMutateAndValidation{}
	resources.Register(r, d)
}

type daemonsetMutateAndValidation struct{}

func (d *daemonsetMutateAndValidation) Mutate(newObject interface{}, raw []byte) (patchBytes []byte, change bool, err error) {
	newDaemonset, ok := newObject.(*appsv1.DaemonSet)
	if !ok {
		err = fmt.Errorf("interface should be daemonset")
		return
	}
	klog.Infof("muta..............%s", newDaemonset.Namespace)
	newDaemonset, err = d.mutate(newDaemonset)
	if err != nil {
		return
	}

	// create json patch
	encoder := scheme.Codecs.LegacyCodec(appsv1.SchemeGroupVersion)
	var modified bytes.Buffer

	err = encoder.Encode(newDaemonset, &modified)
	if err != nil {
		return
	}
	var patch []jsonpatch.JsonPatchOperation
	patch, err = jsonpatch.CreatePatch(raw, modified.Bytes())
	if err != nil {
		return
	}
	patchBytes, err = json.Marshal(patch)
	if err != nil {
		return
	}
	change = true
	return
}

func (d *daemonsetMutateAndValidation) Validate(request *admissionv1beta1.AdmissionRequest) (newObject, oldObject interface{}, err error) {
	raw := request.Object.Raw
	newDaemonset := &appsv1.DaemonSet{}
	oldDaemonset := &appsv1.DaemonSet{}
	deserializer := scheme.Codecs.UniversalDeserializer()
	_, _, err = deserializer.Decode(raw, nil, newDaemonset)
	if err != nil {
		return
	}
	// The daemonset struct loses namespace field when perform update operator,
	// so we need replenish it from request's field.
	newDaemonset.Namespace = request.Namespace
	if request.Operation == admissionv1beta1.Update {
		_, _, err = deserializer.Decode(request.OldObject.Raw, nil, oldDaemonset)
		if err != nil {
			return
		}
		oldDaemonset.Namespace = request.Namespace
		// make sure network configuration remain unchanged
		oldNetworkConfiguration, oldOk := oldDaemonset.Spec.Template.Annotations[containerNetworkconfiguration]
		newNetworkConfiguration, newOk := newDaemonset.Spec.Template.Annotations[containerNetworkconfiguration]
		if oldOk != newOk || oldNetworkConfiguration != newNetworkConfiguration {
			err = fmt.Errorf("network configuration should be constant")
			return
		}
	}
	if newDaemonset.Spec.Selector == nil {
		err = fmt.Errorf("Spec.Selector can't be empty")
		return
	}
	newObject = newDaemonset
	oldObject = oldDaemonset
	return
}

func (d *daemonsetMutateAndValidation) mutate(daemonset *appsv1.DaemonSet) (*appsv1.DaemonSet, error) {
	m := &resources.MutateTools{
		Namespace:   daemonset.Namespace,
		ReleaseName: daemonset.Name,
		Kind:        "daemonset",
	}
	daemonset.Annotations = resources.MutateAnnotations(m, daemonset.Annotations)
	daemonset.Labels = resources.MutateLabels(m, daemonset.Labels)
	daemonset.Spec.Selector.MatchLabels = resources.MutateMatchLabels(m, daemonset.Spec.Selector.MatchLabels)
	resources.MutateTemplate(m, &daemonset.Spec.Template)
	return daemonset, nil
}
