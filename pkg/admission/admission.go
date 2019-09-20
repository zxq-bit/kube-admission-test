package admission

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/resources"

	"github.com/caicloud/clientset/informers"
	"github.com/caicloud/clientset/kubernetes"
	"github.com/caicloud/clientset/kubernetes/scheme"
	apiextensions "github.com/caicloud/clientset/pkg/apis/apiextensions/v1beta1"
	workloadapi "github.com/caicloud/clientset/pkg/apis/workload/v1alpha1"
	kubeutils "github.com/caicloud/go-common/kubernetes"

	"k8s.io/api/admission/v1beta1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const (
	webhookServiceName = "webhook.service"
)

// Server is server for admission
type Server struct {
	client      kubernetes.Interface
	factory     informers.SharedInformerFactory
	tlsProvider *tlsProvider
}

// NewServer returns a new Server
func NewServer(client kubernetes.Interface) (*Server, error) {
	factory := informers.NewSharedInformerFactory(client, 1*time.Hour)
	klog.Infof("new Server...")
	return &Server{
		client:      client,
		factory:     factory,
		tlsProvider: newTLSProvider(client),
	}, nil
}

// Run running
func (s *Server) Run(stopCh <-chan struct{}, port int) error {
	// get key and cert
	tlsConfig, caPEM, err := s.tlsProvider.retrieveTLSConfig()
	if err != nil {
		return err
	}

	if err := s.ensureWebhook(caPEM); err != nil {
		klog.Errorf("Error ensuring webhook, %v", err)
		return err
	}

	if err := s.ensureWorkloadCRD(); err != nil {
		klog.Errorf("Error ensuring workload crd, %v", err)
		return err
	}
	// run informers
	klog.Info("Startting informer factory")
	s.factory.Start(stopCh)

	klog.Info("Waiting for caches synced")
	synced := s.factory.WaitForCacheSync(stopCh)
	for tpy, sync := range synced {
		if !sync {
			err := fmt.Errorf("Wait for %v cache sync timeout", tpy)
			klog.Error(err)
			return err
		}
	}

	klog.Info("All caches have synced!")
	klog.Info("Startting admission server")

	// register handler
	http.HandleFunc("/", s.serve)

	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", port),
		TLSConfig: tlsConfig,
	}
	defer server.Close()

	go func() {
		klog.Infof("Lisen and serve tls on %d", port)
		err := server.ListenAndServeTLS("", "")
		if err != nil {
			klog.Error("Error listenAndServe tls")
		}
	}()

	<-stopCh

	return nil
}
func (s *Server) ensureWorkloadCRD() error {
	crds := []*apiextensions.CustomResourceDefinition{
		&apiextensions.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s.%s", workloadapi.WorkloadPlural, workloadapi.GroupName),
			},
			Spec: apiextensions.CustomResourceDefinitionSpec{
				Group:   workloadapi.GroupName,
				Version: workloadapi.SchemeGroupVersion.Version,
				Scope:   apiextensions.NamespaceScoped,
				Names: apiextensions.CustomResourceDefinitionNames{
					Plural:   workloadapi.WorkloadPlural,
					Singular: "workload",
					Kind:     "Workload",
					ListKind: "WorkloadList",
				},
			},
		},
		&apiextensions.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s.%s", workloadapi.WorkloadRevisionPlural, workloadapi.GroupName),
			},
			Spec: apiextensions.CustomResourceDefinitionSpec{
				Group:   workloadapi.GroupName,
				Version: workloadapi.SchemeGroupVersion.Version,
				Scope:   apiextensions.NamespaceScoped,
				Names: apiextensions.CustomResourceDefinitionNames{
					Plural:   workloadapi.WorkloadRevisionPlural,
					Singular: "workloadrevision",
					Kind:     "WorkloadRevision",
					ListKind: "WorkloadRevisionList",
				},
			},
		},
	}

	for _, crd := range crds {
		incluster, err := s.client.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crd.Name, metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
		if errors.IsNotFound(err) {
			// create
			_, err = s.client.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
			if err != nil {
				klog.Errorf("Failted to create CustomResourceDefinition %v, %v", crd.Name, err)
				return err
			}
			klog.Infof("Create CustomResourceDefinition %v successfully", crd.Name)
			continue
		}
		// update
		incluster.Spec = crd.Spec
		_, err = s.client.ApiextensionsV1beta1().CustomResourceDefinitions().Update(incluster)
		if err != nil {
			klog.Errorf("Failted to update CustomResourceDefinition %v, %v", crd.Name, err)
			return err
		}
		klog.Infof("Update CustomResourceDefinition %v successfully", crd.Name)
	}
	return nil
}

func (s *Server) ensureWebhook(caPem []byte) error {
	fail := admissionregistrationv1beta1.Fail
	serviceName := webhookServiceName
	// webhook := &admissionregistrationv1beta1.MutatingWebhookConfiguration{
	var rules admissionregistrationv1beta1.Rule
	for _, gvr := range resources.GetGVRList() {
		rules.APIGroups = append(rules.APIGroups, gvr.Group)
		rules.APIVersions = append(rules.APIVersions, gvr.Version)
		rules.Resources = append(rules.Resources, gvr.Resource)
	}

	webhook := &admissionregistrationv1beta1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "workload-mutating",
		},
		Webhooks: []admissionregistrationv1beta1.Webhook{
			{
				Name: fmt.Sprintf("%s.%s", "workload", "workload.caicloud.io"),
				Rules: []admissionregistrationv1beta1.RuleWithOperations{
					{
						Operations: []admissionregistrationv1beta1.OperationType{
							admissionregistrationv1beta1.Create,
							admissionregistrationv1beta1.Update,
						},
						Rule: rules,
					},
				},
				ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
					Service: &admissionregistrationv1beta1.ServiceReference{
						Name:      serviceName,
						Namespace: metav1.NamespaceSystem,
					},
					CABundle: caPem,
				},
				FailurePolicy: &fail,
			},
		},
	}

	incluster, err := s.client.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Get(webhook.Name, metav1.GetOptions{})
	// not found: create it
	if errors.IsNotFound(err) {
		_, err = s.client.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Create(webhook)
		if err != nil {
			klog.Errorf("Failted to create MutatingWebhookConfigurations, %v", err)
			return err
		}
		klog.Info("Create MutatingWebhookConfigurations successfully")
		return nil
	}
	// found: update it
	if err == nil {
		incluster.Webhooks = webhook.Webhooks
		_, err = s.client.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Update(incluster)
		if err != nil {
			klog.Errorf("Failted to update MutatingWebhookConfigurations, %v", err)
			return err
		}
		klog.Info("Update MutatingWebhookConfigurations successfully")
		return nil
	}
	// error: return error
	return err
}

func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	klog.Info("recv info ...")
	// 1. verify content type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expected: application/json", contentType)
		return
	}
	// 2. read data from request body
	if r.Body == nil {
		klog.Errorf("request body is empty")
		return
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		klog.Error(err)
		return
	}
	// 3. decode data
	var reviewResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	deserializer := scheme.Codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(data, nil, &ar); err != nil {
		klog.Error(err)
		reviewResponse = toAdmissionResponse(err)
	} else {
		reviewResponse = s.admit(ar)
	}
	// 4. write result to response
	response := v1beta1.AdmissionReview{}
	if reviewResponse != nil {
		response.Response = reviewResponse
		response.Response.UID = ar.Request.UID
	}

	resp, err := json.Marshal(response)
	if err != nil {
		klog.Error(err)
	}
	if _, err := w.Write(resp); err != nil {
		klog.Error(err)
	}
}

func (s *Server) admit(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	reviewResponse := &v1beta1.AdmissionResponse{
		Allowed: true,
	}
	// if namespace should be ignored, just return
	if kubeutils.IgnoredNamespace(ar.Request.Namespace) {
		return reviewResponse
	}
	klog.Infof("admit...%s", string(ar.Request.Object.Raw))
	rmv, ok := resources.GetResourceMutateAndValidation(ar.Request.Resource)
	if !ok {
		klog.Errorf("resource can't be find: %s", ar.Request.Resource)
		return toAdmissionResponse(fmt.Errorf("resource can't be find: %s", ar.Request.Resource))
	}

	newResource, _, err := rmv.Validate(ar.Request)
	if err != nil {
		klog.Error(err)
		return toAdmissionResponse(err)
	}

	patchBytes, change, err := rmv.Mutate(newResource, ar.Request.Object.Raw)
	if err != nil {
		klog.Error(err)
		return toAdmissionResponse(err)
	}
	if !change {
		klog.Infof("Skip validating %v/%v because there is no change", ar.Request.Namespace, ar.Request.Name)
		return reviewResponse
	}
	reviewResponse.Patch = patchBytes
	pt := v1beta1.PatchTypeJSONPatch
	reviewResponse.PatchType = &pt

	klog.Infof("after mutating, the patch is %v", string(patchBytes))
	return reviewResponse
}
