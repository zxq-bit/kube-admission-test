package gen

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/errors"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/review"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/review/handler"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/review/processor"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/util/middlewares/tracer"

	wklv1a1 "github.com/caicloud/clientset/pkg/apis/workload/v1alpha1"
	"github.com/caicloud/go-common/interfaces"
	"github.com/caicloud/nirvana/log"

	arv1b1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func init() {
	review.RegisterHandlerMaker(podsGVR, NewPodHandler)
	review.RegisterHandlerMaker(configmapsGVR, NewConfigMapHandler)
	review.RegisterHandlerMaker(secretsGVR, NewSecretHandler)
	review.RegisterHandlerMaker(servicesGVR, NewServiceHandler)
	review.RegisterHandlerMaker(persistentvolumeclaimsGVR, NewPersistentVolumeClaimHandler)
	review.RegisterHandlerMaker(persistentvolumesGVR, NewPersistentVolumeHandler)
	review.RegisterHandlerMaker(daemonsetsGVR, NewDaemonSetHandler)
	review.RegisterHandlerMaker(deploymentsGVR, NewDeploymentHandler)
	review.RegisterHandlerMaker(replicasetsGVR, NewReplicaSetHandler)
	review.RegisterHandlerMaker(statefulsetsGVR, NewStatefulSetHandler)
	review.RegisterHandlerMaker(workloadsGVR, NewWorkloadHandler)
}

var (
	podsGVR                   = corev1.SchemeGroupVersion.WithResource("pods")
	podsGVK                   = corev1.SchemeGroupVersion.WithKind("Pod")
	configmapsGVR             = corev1.SchemeGroupVersion.WithResource("configmaps")
	configmapsGVK             = corev1.SchemeGroupVersion.WithKind("ConfigMap")
	secretsGVR                = corev1.SchemeGroupVersion.WithResource("secrets")
	secretsGVK                = corev1.SchemeGroupVersion.WithKind("Secret")
	servicesGVR               = corev1.SchemeGroupVersion.WithResource("services")
	servicesGVK               = corev1.SchemeGroupVersion.WithKind("Service")
	persistentvolumeclaimsGVR = corev1.SchemeGroupVersion.WithResource("persistentvolumeclaims")
	persistentvolumeclaimsGVK = corev1.SchemeGroupVersion.WithKind("PersistentVolumeClaim")
	persistentvolumesGVR      = corev1.SchemeGroupVersion.WithResource("persistentvolumes")
	persistentvolumesGVK      = corev1.SchemeGroupVersion.WithKind("PersistentVolume")
	daemonsetsGVR             = appsv1.SchemeGroupVersion.WithResource("daemonsets")
	daemonsetsGVK             = appsv1.SchemeGroupVersion.WithKind("DaemonSet")
	deploymentsGVR            = appsv1.SchemeGroupVersion.WithResource("deployments")
	deploymentsGVK            = appsv1.SchemeGroupVersion.WithKind("Deployment")
	replicasetsGVR            = appsv1.SchemeGroupVersion.WithResource("replicasets")
	replicasetsGVK            = appsv1.SchemeGroupVersion.WithKind("ReplicaSet")
	statefulsetsGVR           = appsv1.SchemeGroupVersion.WithResource("statefulsets")
	statefulsetsGVK           = appsv1.SchemeGroupVersion.WithKind("StatefulSet")
	workloadsGVR              = wklv1a1.SchemeGroupVersion.WithResource("workloads")
	workloadsGVK              = wklv1a1.SchemeGroupVersion.WithKind("Workload")

	gvr2gvkMap = map[schema.GroupVersionResource]schema.GroupVersionKind{
		podsGVR:                   podsGVK,
		configmapsGVR:             configmapsGVK,
		secretsGVR:                secretsGVK,
		servicesGVR:               servicesGVK,
		persistentvolumeclaimsGVR: persistentvolumeclaimsGVK,
		persistentvolumesGVR:      persistentvolumesGVK,
		daemonsetsGVR:             daemonsetsGVK,
		deploymentsGVR:            deploymentsGVK,
		replicasetsGVR:            replicasetsGVK,
		statefulsetsGVR:           statefulsetsGVK,
		workloadsGVR:              workloadsGVK,
	}
	gvk2gvrMap = map[schema.GroupVersionKind]schema.GroupVersionResource{
		podsGVK:                   podsGVR,
		configmapsGVK:             configmapsGVR,
		secretsGVK:                secretsGVR,
		servicesGVK:               servicesGVR,
		persistentvolumeclaimsGVK: persistentvolumeclaimsGVR,
		persistentvolumesGVK:      persistentvolumesGVR,
		daemonsetsGVK:             daemonsetsGVR,
		deploymentsGVK:            deploymentsGVR,
		replicasetsGVK:            replicasetsGVR,
		statefulsetsGVK:           statefulsetsGVR,
		workloadsGVK:              workloadsGVR,
	}
)

func PodGVR() schema.GroupVersionResource                   { return podsGVR }
func PodGVK() schema.GroupVersionKind                       { return podsGVK }
func ConfigMapGVR() schema.GroupVersionResource             { return configmapsGVR }
func ConfigMapGVK() schema.GroupVersionKind                 { return configmapsGVK }
func SecretGVR() schema.GroupVersionResource                { return secretsGVR }
func SecretGVK() schema.GroupVersionKind                    { return secretsGVK }
func ServiceGVR() schema.GroupVersionResource               { return servicesGVR }
func ServiceGVK() schema.GroupVersionKind                   { return servicesGVK }
func PersistentVolumeClaimGVR() schema.GroupVersionResource { return persistentvolumeclaimsGVR }
func PersistentVolumeClaimGVK() schema.GroupVersionKind     { return persistentvolumeclaimsGVK }
func PersistentVolumeGVR() schema.GroupVersionResource      { return persistentvolumesGVR }
func PersistentVolumeGVK() schema.GroupVersionKind          { return persistentvolumesGVK }
func DaemonSetGVR() schema.GroupVersionResource             { return daemonsetsGVR }
func DaemonSetGVK() schema.GroupVersionKind                 { return daemonsetsGVK }
func DeploymentGVR() schema.GroupVersionResource            { return deploymentsGVR }
func DeploymentGVK() schema.GroupVersionKind                { return deploymentsGVK }
func ReplicaSetGVR() schema.GroupVersionResource            { return replicasetsGVR }
func ReplicaSetGVK() schema.GroupVersionKind                { return replicasetsGVK }
func StatefulSetGVR() schema.GroupVersionResource           { return statefulsetsGVR }
func StatefulSetGVK() schema.GroupVersionKind               { return statefulsetsGVK }
func WorkloadGVR() schema.GroupVersionResource              { return workloadsGVR }
func WorkloadGVK() schema.GroupVersionKind                  { return workloadsGVK }

func GetGVKByGVR(in schema.GroupVersionResource) (out schema.GroupVersionKind, ok bool) {
	out, ok = gvr2gvkMap[in]
	return
}

func GetGVRByGVK(in schema.GroupVersionKind) (out schema.GroupVersionResource, ok bool) {
	out, ok = gvk2gvrMap[in]
	return
}

// pods about

type PodProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Tracer, do performance tracking
	Tracer tracer.Tracer
	// Admit do admit, return error if should stop
	Admit func(ctx context.Context, in *corev1.Pod) errors.APIStatus
}

type PodHandler struct {
	processors []*PodProcessor
	objFilters []util.ObjectIgnoreFilter
}

// pods processor

func (p *PodProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Admit == nil {
		return fmt.Errorf("%v nil processor admit function", p.Key())
	}
	return nil
}

func (p *PodProcessor) DoWithTracing(ctx context.Context, in *corev1.Pod) (cost time.Duration, ke errors.APIStatus) {
	return p.Tracer.DoWithTracing(func() errors.APIStatus {
		return p.Admit(ctx, in)
	})
}

// pods handler

func NewPodHandler(opType arv1b1.OperationType) (review.Handler, error) {
	return handler.NewFramework(
		podsGVR,
		opType,
		func(raw *runtime.RawExtension) (runtime.Object, error) {
			return podsRawExtensionParser(raw)
		},
		&PodHandler{},
	)
}

func (h *PodHandler) IsEmpty() bool {
	return len(h.processors) == 0
}

func (h *PodHandler) Register(in interface{}) error {
	getProcessor := func(v interface{}) *PodProcessor {
		if v == nil {
			return nil
		}
		return v.(*PodProcessor)
	}
	p := getProcessor(in)
	if p == nil {
		return errors.ErrProcessorIsNil
	}
	if e := p.Validate(); e != nil {
		return e
	}
	h.processors = append(h.processors, p)
	h.objFilters = append(h.objFilters, p.GetObjectFilter())
	return nil
}

func (h *PodHandler) DoAdmit(ctx context.Context, tracer *tracer.Tracer, in runtime.Object) (cost time.Duration, ke errors.APIStatus) {
	return tracer.DoWithTracing(func() (ke errors.APIStatus) {
		// log prepare
		logBase := util.GetContextLogBase(ctx)
		// check
		obj := func() *corev1.Pod {
			if interfaces.IsNil(in) {
				return nil
			}
			return in.(*corev1.Pod)
		}()
		toFilter := obj
		if toFilter == nil {
			var err error
			toFilter, err = GetContextOldPod(ctx)
			if err != nil {
				err = errors.ErrWrongRuntimeObjects
				log.Errorf("%s DoAdmit failed, %v", logBase, err)
				return errors.NewBadRequest(err)
			}
		}
		// cleanup
		defer util.RemoveObjectAnno(obj, constants.AnnoKeyAdmissionIgnore)
		// execute processors
		for i, p := range h.processors {
			logPrefix := logBase + fmt.Sprintf("[%d][%s]", i, p.LogPrefix())
			// check ignore
			if ignoreReason := h.objFilters[i](toFilter); ignoreReason != nil {
				log.Infof("%s skip for %s", logPrefix, *ignoreReason)
				continue
			}
			// do review
			cost := time.Duration(0)
			select {
			case <-ctx.Done():
				ke = errors.NewRequestTimeout(errors.ErrContextEnded)
			default:
				switch p.Type {
				case constants.ProcessorTypeValidate: // do without changes
					var toValidate *corev1.Pod
					if obj != nil {
						toValidate = obj.DeepCopy()
					}
					cost, ke = p.DoWithTracing(ctx, toValidate)
				case constants.ProcessorTypeMutate:
					cost, ke = p.DoWithTracing(ctx, obj)
				default:
					log.Errorf("%s skip for unknown processor type '%v'", p.Type)
				}
			}
			if ke != nil {
				log.Errorf("%s[cost:%v] stop by error: %v", logPrefix, cost, ke)
				break
			}
			log.Infof("%s[cost:%v] done", logPrefix, cost)
		}
		return
	})
}

func podsRawExtensionParser(raw *runtime.RawExtension) (*corev1.Pod, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if !interfaces.IsNil(raw.Object) {
		if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != podsGVK {
			return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), podsGVK.String())
		}
		if obj := raw.Object.(*corev1.Pod); obj != nil {
			return obj.DeepCopy(), nil
		}
	}
	parsed := &corev1.Pod{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}

// GetContextOldPod get Pod old object from Context
// no error if old object not exist, error if parse failed
func GetContextOldPod(ctx context.Context) (*corev1.Pod, error) {
	raw := util.GetContextOldObject(ctx)
	if raw == nil { // no old object
		return nil, nil
	}
	return podsRawExtensionParser(raw)
}

// configmaps about

type ConfigMapProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Tracer, do performance tracking
	Tracer tracer.Tracer
	// Admit do admit, return error if should stop
	Admit func(ctx context.Context, in *corev1.ConfigMap) errors.APIStatus
}

type ConfigMapHandler struct {
	processors []*ConfigMapProcessor
	objFilters []util.ObjectIgnoreFilter
}

// configmaps processor

func (p *ConfigMapProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Admit == nil {
		return fmt.Errorf("%v nil processor admit function", p.Key())
	}
	return nil
}

func (p *ConfigMapProcessor) DoWithTracing(ctx context.Context, in *corev1.ConfigMap) (cost time.Duration, ke errors.APIStatus) {
	return p.Tracer.DoWithTracing(func() errors.APIStatus {
		return p.Admit(ctx, in)
	})
}

// configmaps handler

func NewConfigMapHandler(opType arv1b1.OperationType) (review.Handler, error) {
	return handler.NewFramework(
		configmapsGVR,
		opType,
		func(raw *runtime.RawExtension) (runtime.Object, error) {
			return configmapsRawExtensionParser(raw)
		},
		&ConfigMapHandler{},
	)
}

func (h *ConfigMapHandler) IsEmpty() bool {
	return len(h.processors) == 0
}

func (h *ConfigMapHandler) Register(in interface{}) error {
	getProcessor := func(v interface{}) *ConfigMapProcessor {
		if v == nil {
			return nil
		}
		return v.(*ConfigMapProcessor)
	}
	p := getProcessor(in)
	if p == nil {
		return errors.ErrProcessorIsNil
	}
	if e := p.Validate(); e != nil {
		return e
	}
	h.processors = append(h.processors, p)
	h.objFilters = append(h.objFilters, p.GetObjectFilter())
	return nil
}

func (h *ConfigMapHandler) DoAdmit(ctx context.Context, tracer *tracer.Tracer, in runtime.Object) (cost time.Duration, ke errors.APIStatus) {
	return tracer.DoWithTracing(func() (ke errors.APIStatus) {
		// log prepare
		logBase := util.GetContextLogBase(ctx)
		// check
		obj := func() *corev1.ConfigMap {
			if interfaces.IsNil(in) {
				return nil
			}
			return in.(*corev1.ConfigMap)
		}()
		toFilter := obj
		if toFilter == nil {
			var err error
			toFilter, err = GetContextOldConfigMap(ctx)
			if err != nil {
				err = errors.ErrWrongRuntimeObjects
				log.Errorf("%s DoAdmit failed, %v", logBase, err)
				return errors.NewBadRequest(err)
			}
		}
		// cleanup
		defer util.RemoveObjectAnno(obj, constants.AnnoKeyAdmissionIgnore)
		// execute processors
		for i, p := range h.processors {
			logPrefix := logBase + fmt.Sprintf("[%d][%s]", i, p.LogPrefix())
			// check ignore
			if ignoreReason := h.objFilters[i](toFilter); ignoreReason != nil {
				log.Infof("%s skip for %s", logPrefix, *ignoreReason)
				continue
			}
			// do review
			cost := time.Duration(0)
			select {
			case <-ctx.Done():
				ke = errors.NewRequestTimeout(errors.ErrContextEnded)
			default:
				switch p.Type {
				case constants.ProcessorTypeValidate: // do without changes
					var toValidate *corev1.ConfigMap
					if obj != nil {
						toValidate = obj.DeepCopy()
					}
					cost, ke = p.DoWithTracing(ctx, toValidate)
				case constants.ProcessorTypeMutate:
					cost, ke = p.DoWithTracing(ctx, obj)
				default:
					log.Errorf("%s skip for unknown processor type '%v'", p.Type)
				}
			}
			if ke != nil {
				log.Errorf("%s[cost:%v] stop by error: %v", logPrefix, cost, ke)
				break
			}
			log.Infof("%s[cost:%v] done", logPrefix, cost)
		}
		return
	})
}

func configmapsRawExtensionParser(raw *runtime.RawExtension) (*corev1.ConfigMap, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if !interfaces.IsNil(raw.Object) {
		if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != configmapsGVK {
			return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), configmapsGVK.String())
		}
		if obj := raw.Object.(*corev1.ConfigMap); obj != nil {
			return obj.DeepCopy(), nil
		}
	}
	parsed := &corev1.ConfigMap{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}

// GetContextOldConfigMap get ConfigMap old object from Context
// no error if old object not exist, error if parse failed
func GetContextOldConfigMap(ctx context.Context) (*corev1.ConfigMap, error) {
	raw := util.GetContextOldObject(ctx)
	if raw == nil { // no old object
		return nil, nil
	}
	return configmapsRawExtensionParser(raw)
}

// secrets about

type SecretProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Tracer, do performance tracking
	Tracer tracer.Tracer
	// Admit do admit, return error if should stop
	Admit func(ctx context.Context, in *corev1.Secret) errors.APIStatus
}

type SecretHandler struct {
	processors []*SecretProcessor
	objFilters []util.ObjectIgnoreFilter
}

// secrets processor

func (p *SecretProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Admit == nil {
		return fmt.Errorf("%v nil processor admit function", p.Key())
	}
	return nil
}

func (p *SecretProcessor) DoWithTracing(ctx context.Context, in *corev1.Secret) (cost time.Duration, ke errors.APIStatus) {
	return p.Tracer.DoWithTracing(func() errors.APIStatus {
		return p.Admit(ctx, in)
	})
}

// secrets handler

func NewSecretHandler(opType arv1b1.OperationType) (review.Handler, error) {
	return handler.NewFramework(
		secretsGVR,
		opType,
		func(raw *runtime.RawExtension) (runtime.Object, error) {
			return secretsRawExtensionParser(raw)
		},
		&SecretHandler{},
	)
}

func (h *SecretHandler) IsEmpty() bool {
	return len(h.processors) == 0
}

func (h *SecretHandler) Register(in interface{}) error {
	getProcessor := func(v interface{}) *SecretProcessor {
		if v == nil {
			return nil
		}
		return v.(*SecretProcessor)
	}
	p := getProcessor(in)
	if p == nil {
		return errors.ErrProcessorIsNil
	}
	if e := p.Validate(); e != nil {
		return e
	}
	h.processors = append(h.processors, p)
	h.objFilters = append(h.objFilters, p.GetObjectFilter())
	return nil
}

func (h *SecretHandler) DoAdmit(ctx context.Context, tracer *tracer.Tracer, in runtime.Object) (cost time.Duration, ke errors.APIStatus) {
	return tracer.DoWithTracing(func() (ke errors.APIStatus) {
		// log prepare
		logBase := util.GetContextLogBase(ctx)
		// check
		obj := func() *corev1.Secret {
			if interfaces.IsNil(in) {
				return nil
			}
			return in.(*corev1.Secret)
		}()
		toFilter := obj
		if toFilter == nil {
			var err error
			toFilter, err = GetContextOldSecret(ctx)
			if err != nil {
				err = errors.ErrWrongRuntimeObjects
				log.Errorf("%s DoAdmit failed, %v", logBase, err)
				return errors.NewBadRequest(err)
			}
		}
		// cleanup
		defer util.RemoveObjectAnno(obj, constants.AnnoKeyAdmissionIgnore)
		// execute processors
		for i, p := range h.processors {
			logPrefix := logBase + fmt.Sprintf("[%d][%s]", i, p.LogPrefix())
			// check ignore
			if ignoreReason := h.objFilters[i](toFilter); ignoreReason != nil {
				log.Infof("%s skip for %s", logPrefix, *ignoreReason)
				continue
			}
			// do review
			cost := time.Duration(0)
			select {
			case <-ctx.Done():
				ke = errors.NewRequestTimeout(errors.ErrContextEnded)
			default:
				switch p.Type {
				case constants.ProcessorTypeValidate: // do without changes
					var toValidate *corev1.Secret
					if obj != nil {
						toValidate = obj.DeepCopy()
					}
					cost, ke = p.DoWithTracing(ctx, toValidate)
				case constants.ProcessorTypeMutate:
					cost, ke = p.DoWithTracing(ctx, obj)
				default:
					log.Errorf("%s skip for unknown processor type '%v'", p.Type)
				}
			}
			if ke != nil {
				log.Errorf("%s[cost:%v] stop by error: %v", logPrefix, cost, ke)
				break
			}
			log.Infof("%s[cost:%v] done", logPrefix, cost)
		}
		return
	})
}

func secretsRawExtensionParser(raw *runtime.RawExtension) (*corev1.Secret, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if !interfaces.IsNil(raw.Object) {
		if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != secretsGVK {
			return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), secretsGVK.String())
		}
		if obj := raw.Object.(*corev1.Secret); obj != nil {
			return obj.DeepCopy(), nil
		}
	}
	parsed := &corev1.Secret{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}

// GetContextOldSecret get Secret old object from Context
// no error if old object not exist, error if parse failed
func GetContextOldSecret(ctx context.Context) (*corev1.Secret, error) {
	raw := util.GetContextOldObject(ctx)
	if raw == nil { // no old object
		return nil, nil
	}
	return secretsRawExtensionParser(raw)
}

// services about

type ServiceProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Tracer, do performance tracking
	Tracer tracer.Tracer
	// Admit do admit, return error if should stop
	Admit func(ctx context.Context, in *corev1.Service) errors.APIStatus
}

type ServiceHandler struct {
	processors []*ServiceProcessor
	objFilters []util.ObjectIgnoreFilter
}

// services processor

func (p *ServiceProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Admit == nil {
		return fmt.Errorf("%v nil processor admit function", p.Key())
	}
	return nil
}

func (p *ServiceProcessor) DoWithTracing(ctx context.Context, in *corev1.Service) (cost time.Duration, ke errors.APIStatus) {
	return p.Tracer.DoWithTracing(func() errors.APIStatus {
		return p.Admit(ctx, in)
	})
}

// services handler

func NewServiceHandler(opType arv1b1.OperationType) (review.Handler, error) {
	return handler.NewFramework(
		servicesGVR,
		opType,
		func(raw *runtime.RawExtension) (runtime.Object, error) {
			return servicesRawExtensionParser(raw)
		},
		&ServiceHandler{},
	)
}

func (h *ServiceHandler) IsEmpty() bool {
	return len(h.processors) == 0
}

func (h *ServiceHandler) Register(in interface{}) error {
	getProcessor := func(v interface{}) *ServiceProcessor {
		if v == nil {
			return nil
		}
		return v.(*ServiceProcessor)
	}
	p := getProcessor(in)
	if p == nil {
		return errors.ErrProcessorIsNil
	}
	if e := p.Validate(); e != nil {
		return e
	}
	h.processors = append(h.processors, p)
	h.objFilters = append(h.objFilters, p.GetObjectFilter())
	return nil
}

func (h *ServiceHandler) DoAdmit(ctx context.Context, tracer *tracer.Tracer, in runtime.Object) (cost time.Duration, ke errors.APIStatus) {
	return tracer.DoWithTracing(func() (ke errors.APIStatus) {
		// log prepare
		logBase := util.GetContextLogBase(ctx)
		// check
		obj := func() *corev1.Service {
			if interfaces.IsNil(in) {
				return nil
			}
			return in.(*corev1.Service)
		}()
		toFilter := obj
		if toFilter == nil {
			var err error
			toFilter, err = GetContextOldService(ctx)
			if err != nil {
				err = errors.ErrWrongRuntimeObjects
				log.Errorf("%s DoAdmit failed, %v", logBase, err)
				return errors.NewBadRequest(err)
			}
		}
		// cleanup
		defer util.RemoveObjectAnno(obj, constants.AnnoKeyAdmissionIgnore)
		// execute processors
		for i, p := range h.processors {
			logPrefix := logBase + fmt.Sprintf("[%d][%s]", i, p.LogPrefix())
			// check ignore
			if ignoreReason := h.objFilters[i](toFilter); ignoreReason != nil {
				log.Infof("%s skip for %s", logPrefix, *ignoreReason)
				continue
			}
			// do review
			cost := time.Duration(0)
			select {
			case <-ctx.Done():
				ke = errors.NewRequestTimeout(errors.ErrContextEnded)
			default:
				switch p.Type {
				case constants.ProcessorTypeValidate: // do without changes
					var toValidate *corev1.Service
					if obj != nil {
						toValidate = obj.DeepCopy()
					}
					cost, ke = p.DoWithTracing(ctx, toValidate)
				case constants.ProcessorTypeMutate:
					cost, ke = p.DoWithTracing(ctx, obj)
				default:
					log.Errorf("%s skip for unknown processor type '%v'", p.Type)
				}
			}
			if ke != nil {
				log.Errorf("%s[cost:%v] stop by error: %v", logPrefix, cost, ke)
				break
			}
			log.Infof("%s[cost:%v] done", logPrefix, cost)
		}
		return
	})
}

func servicesRawExtensionParser(raw *runtime.RawExtension) (*corev1.Service, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if !interfaces.IsNil(raw.Object) {
		if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != servicesGVK {
			return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), servicesGVK.String())
		}
		if obj := raw.Object.(*corev1.Service); obj != nil {
			return obj.DeepCopy(), nil
		}
	}
	parsed := &corev1.Service{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}

// GetContextOldService get Service old object from Context
// no error if old object not exist, error if parse failed
func GetContextOldService(ctx context.Context) (*corev1.Service, error) {
	raw := util.GetContextOldObject(ctx)
	if raw == nil { // no old object
		return nil, nil
	}
	return servicesRawExtensionParser(raw)
}

// persistentvolumeclaims about

type PersistentVolumeClaimProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Tracer, do performance tracking
	Tracer tracer.Tracer
	// Admit do admit, return error if should stop
	Admit func(ctx context.Context, in *corev1.PersistentVolumeClaim) errors.APIStatus
}

type PersistentVolumeClaimHandler struct {
	processors []*PersistentVolumeClaimProcessor
	objFilters []util.ObjectIgnoreFilter
}

// persistentvolumeclaims processor

func (p *PersistentVolumeClaimProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Admit == nil {
		return fmt.Errorf("%v nil processor admit function", p.Key())
	}
	return nil
}

func (p *PersistentVolumeClaimProcessor) DoWithTracing(ctx context.Context, in *corev1.PersistentVolumeClaim) (cost time.Duration, ke errors.APIStatus) {
	return p.Tracer.DoWithTracing(func() errors.APIStatus {
		return p.Admit(ctx, in)
	})
}

// persistentvolumeclaims handler

func NewPersistentVolumeClaimHandler(opType arv1b1.OperationType) (review.Handler, error) {
	return handler.NewFramework(
		persistentvolumeclaimsGVR,
		opType,
		func(raw *runtime.RawExtension) (runtime.Object, error) {
			return persistentvolumeclaimsRawExtensionParser(raw)
		},
		&PersistentVolumeClaimHandler{},
	)
}

func (h *PersistentVolumeClaimHandler) IsEmpty() bool {
	return len(h.processors) == 0
}

func (h *PersistentVolumeClaimHandler) Register(in interface{}) error {
	getProcessor := func(v interface{}) *PersistentVolumeClaimProcessor {
		if v == nil {
			return nil
		}
		return v.(*PersistentVolumeClaimProcessor)
	}
	p := getProcessor(in)
	if p == nil {
		return errors.ErrProcessorIsNil
	}
	if e := p.Validate(); e != nil {
		return e
	}
	h.processors = append(h.processors, p)
	h.objFilters = append(h.objFilters, p.GetObjectFilter())
	return nil
}

func (h *PersistentVolumeClaimHandler) DoAdmit(ctx context.Context, tracer *tracer.Tracer, in runtime.Object) (cost time.Duration, ke errors.APIStatus) {
	return tracer.DoWithTracing(func() (ke errors.APIStatus) {
		// log prepare
		logBase := util.GetContextLogBase(ctx)
		// check
		obj := func() *corev1.PersistentVolumeClaim {
			if interfaces.IsNil(in) {
				return nil
			}
			return in.(*corev1.PersistentVolumeClaim)
		}()
		toFilter := obj
		if toFilter == nil {
			var err error
			toFilter, err = GetContextOldPersistentVolumeClaim(ctx)
			if err != nil {
				err = errors.ErrWrongRuntimeObjects
				log.Errorf("%s DoAdmit failed, %v", logBase, err)
				return errors.NewBadRequest(err)
			}
		}
		// cleanup
		defer util.RemoveObjectAnno(obj, constants.AnnoKeyAdmissionIgnore)
		// execute processors
		for i, p := range h.processors {
			logPrefix := logBase + fmt.Sprintf("[%d][%s]", i, p.LogPrefix())
			// check ignore
			if ignoreReason := h.objFilters[i](toFilter); ignoreReason != nil {
				log.Infof("%s skip for %s", logPrefix, *ignoreReason)
				continue
			}
			// do review
			cost := time.Duration(0)
			select {
			case <-ctx.Done():
				ke = errors.NewRequestTimeout(errors.ErrContextEnded)
			default:
				switch p.Type {
				case constants.ProcessorTypeValidate: // do without changes
					var toValidate *corev1.PersistentVolumeClaim
					if obj != nil {
						toValidate = obj.DeepCopy()
					}
					cost, ke = p.DoWithTracing(ctx, toValidate)
				case constants.ProcessorTypeMutate:
					cost, ke = p.DoWithTracing(ctx, obj)
				default:
					log.Errorf("%s skip for unknown processor type '%v'", p.Type)
				}
			}
			if ke != nil {
				log.Errorf("%s[cost:%v] stop by error: %v", logPrefix, cost, ke)
				break
			}
			log.Infof("%s[cost:%v] done", logPrefix, cost)
		}
		return
	})
}

func persistentvolumeclaimsRawExtensionParser(raw *runtime.RawExtension) (*corev1.PersistentVolumeClaim, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if !interfaces.IsNil(raw.Object) {
		if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != persistentvolumeclaimsGVK {
			return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), persistentvolumeclaimsGVK.String())
		}
		if obj := raw.Object.(*corev1.PersistentVolumeClaim); obj != nil {
			return obj.DeepCopy(), nil
		}
	}
	parsed := &corev1.PersistentVolumeClaim{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}

// GetContextOldPersistentVolumeClaim get PersistentVolumeClaim old object from Context
// no error if old object not exist, error if parse failed
func GetContextOldPersistentVolumeClaim(ctx context.Context) (*corev1.PersistentVolumeClaim, error) {
	raw := util.GetContextOldObject(ctx)
	if raw == nil { // no old object
		return nil, nil
	}
	return persistentvolumeclaimsRawExtensionParser(raw)
}

// persistentvolumes about

type PersistentVolumeProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Tracer, do performance tracking
	Tracer tracer.Tracer
	// Admit do admit, return error if should stop
	Admit func(ctx context.Context, in *corev1.PersistentVolume) errors.APIStatus
}

type PersistentVolumeHandler struct {
	processors []*PersistentVolumeProcessor
	objFilters []util.ObjectIgnoreFilter
}

// persistentvolumes processor

func (p *PersistentVolumeProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Admit == nil {
		return fmt.Errorf("%v nil processor admit function", p.Key())
	}
	return nil
}

func (p *PersistentVolumeProcessor) DoWithTracing(ctx context.Context, in *corev1.PersistentVolume) (cost time.Duration, ke errors.APIStatus) {
	return p.Tracer.DoWithTracing(func() errors.APIStatus {
		return p.Admit(ctx, in)
	})
}

// persistentvolumes handler

func NewPersistentVolumeHandler(opType arv1b1.OperationType) (review.Handler, error) {
	return handler.NewFramework(
		persistentvolumesGVR,
		opType,
		func(raw *runtime.RawExtension) (runtime.Object, error) {
			return persistentvolumesRawExtensionParser(raw)
		},
		&PersistentVolumeHandler{},
	)
}

func (h *PersistentVolumeHandler) IsEmpty() bool {
	return len(h.processors) == 0
}

func (h *PersistentVolumeHandler) Register(in interface{}) error {
	getProcessor := func(v interface{}) *PersistentVolumeProcessor {
		if v == nil {
			return nil
		}
		return v.(*PersistentVolumeProcessor)
	}
	p := getProcessor(in)
	if p == nil {
		return errors.ErrProcessorIsNil
	}
	if e := p.Validate(); e != nil {
		return e
	}
	h.processors = append(h.processors, p)
	h.objFilters = append(h.objFilters, p.GetObjectFilter())
	return nil
}

func (h *PersistentVolumeHandler) DoAdmit(ctx context.Context, tracer *tracer.Tracer, in runtime.Object) (cost time.Duration, ke errors.APIStatus) {
	return tracer.DoWithTracing(func() (ke errors.APIStatus) {
		// log prepare
		logBase := util.GetContextLogBase(ctx)
		// check
		obj := func() *corev1.PersistentVolume {
			if interfaces.IsNil(in) {
				return nil
			}
			return in.(*corev1.PersistentVolume)
		}()
		toFilter := obj
		if toFilter == nil {
			var err error
			toFilter, err = GetContextOldPersistentVolume(ctx)
			if err != nil {
				err = errors.ErrWrongRuntimeObjects
				log.Errorf("%s DoAdmit failed, %v", logBase, err)
				return errors.NewBadRequest(err)
			}
		}
		// cleanup
		defer util.RemoveObjectAnno(obj, constants.AnnoKeyAdmissionIgnore)
		// execute processors
		for i, p := range h.processors {
			logPrefix := logBase + fmt.Sprintf("[%d][%s]", i, p.LogPrefix())
			// check ignore
			if ignoreReason := h.objFilters[i](toFilter); ignoreReason != nil {
				log.Infof("%s skip for %s", logPrefix, *ignoreReason)
				continue
			}
			// do review
			cost := time.Duration(0)
			select {
			case <-ctx.Done():
				ke = errors.NewRequestTimeout(errors.ErrContextEnded)
			default:
				switch p.Type {
				case constants.ProcessorTypeValidate: // do without changes
					var toValidate *corev1.PersistentVolume
					if obj != nil {
						toValidate = obj.DeepCopy()
					}
					cost, ke = p.DoWithTracing(ctx, toValidate)
				case constants.ProcessorTypeMutate:
					cost, ke = p.DoWithTracing(ctx, obj)
				default:
					log.Errorf("%s skip for unknown processor type '%v'", p.Type)
				}
			}
			if ke != nil {
				log.Errorf("%s[cost:%v] stop by error: %v", logPrefix, cost, ke)
				break
			}
			log.Infof("%s[cost:%v] done", logPrefix, cost)
		}
		return
	})
}

func persistentvolumesRawExtensionParser(raw *runtime.RawExtension) (*corev1.PersistentVolume, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if !interfaces.IsNil(raw.Object) {
		if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != persistentvolumesGVK {
			return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), persistentvolumesGVK.String())
		}
		if obj := raw.Object.(*corev1.PersistentVolume); obj != nil {
			return obj.DeepCopy(), nil
		}
	}
	parsed := &corev1.PersistentVolume{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}

// GetContextOldPersistentVolume get PersistentVolume old object from Context
// no error if old object not exist, error if parse failed
func GetContextOldPersistentVolume(ctx context.Context) (*corev1.PersistentVolume, error) {
	raw := util.GetContextOldObject(ctx)
	if raw == nil { // no old object
		return nil, nil
	}
	return persistentvolumesRawExtensionParser(raw)
}

// daemonsets about

type DaemonSetProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Tracer, do performance tracking
	Tracer tracer.Tracer
	// Admit do admit, return error if should stop
	Admit func(ctx context.Context, in *appsv1.DaemonSet) errors.APIStatus
}

type DaemonSetHandler struct {
	processors []*DaemonSetProcessor
	objFilters []util.ObjectIgnoreFilter
}

// daemonsets processor

func (p *DaemonSetProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Admit == nil {
		return fmt.Errorf("%v nil processor admit function", p.Key())
	}
	return nil
}

func (p *DaemonSetProcessor) DoWithTracing(ctx context.Context, in *appsv1.DaemonSet) (cost time.Duration, ke errors.APIStatus) {
	return p.Tracer.DoWithTracing(func() errors.APIStatus {
		return p.Admit(ctx, in)
	})
}

// daemonsets handler

func NewDaemonSetHandler(opType arv1b1.OperationType) (review.Handler, error) {
	return handler.NewFramework(
		daemonsetsGVR,
		opType,
		func(raw *runtime.RawExtension) (runtime.Object, error) {
			return daemonsetsRawExtensionParser(raw)
		},
		&DaemonSetHandler{},
	)
}

func (h *DaemonSetHandler) IsEmpty() bool {
	return len(h.processors) == 0
}

func (h *DaemonSetHandler) Register(in interface{}) error {
	getProcessor := func(v interface{}) *DaemonSetProcessor {
		if v == nil {
			return nil
		}
		return v.(*DaemonSetProcessor)
	}
	p := getProcessor(in)
	if p == nil {
		return errors.ErrProcessorIsNil
	}
	if e := p.Validate(); e != nil {
		return e
	}
	h.processors = append(h.processors, p)
	h.objFilters = append(h.objFilters, p.GetObjectFilter())
	return nil
}

func (h *DaemonSetHandler) DoAdmit(ctx context.Context, tracer *tracer.Tracer, in runtime.Object) (cost time.Duration, ke errors.APIStatus) {
	return tracer.DoWithTracing(func() (ke errors.APIStatus) {
		// log prepare
		logBase := util.GetContextLogBase(ctx)
		// check
		obj := func() *appsv1.DaemonSet {
			if interfaces.IsNil(in) {
				return nil
			}
			return in.(*appsv1.DaemonSet)
		}()
		toFilter := obj
		if toFilter == nil {
			var err error
			toFilter, err = GetContextOldDaemonSet(ctx)
			if err != nil {
				err = errors.ErrWrongRuntimeObjects
				log.Errorf("%s DoAdmit failed, %v", logBase, err)
				return errors.NewBadRequest(err)
			}
		}
		// cleanup
		defer util.RemoveObjectAnno(obj, constants.AnnoKeyAdmissionIgnore)
		// execute processors
		for i, p := range h.processors {
			logPrefix := logBase + fmt.Sprintf("[%d][%s]", i, p.LogPrefix())
			// check ignore
			if ignoreReason := h.objFilters[i](toFilter); ignoreReason != nil {
				log.Infof("%s skip for %s", logPrefix, *ignoreReason)
				continue
			}
			// do review
			cost := time.Duration(0)
			select {
			case <-ctx.Done():
				ke = errors.NewRequestTimeout(errors.ErrContextEnded)
			default:
				switch p.Type {
				case constants.ProcessorTypeValidate: // do without changes
					var toValidate *appsv1.DaemonSet
					if obj != nil {
						toValidate = obj.DeepCopy()
					}
					cost, ke = p.DoWithTracing(ctx, toValidate)
				case constants.ProcessorTypeMutate:
					cost, ke = p.DoWithTracing(ctx, obj)
				default:
					log.Errorf("%s skip for unknown processor type '%v'", p.Type)
				}
			}
			if ke != nil {
				log.Errorf("%s[cost:%v] stop by error: %v", logPrefix, cost, ke)
				break
			}
			log.Infof("%s[cost:%v] done", logPrefix, cost)
		}
		return
	})
}

func daemonsetsRawExtensionParser(raw *runtime.RawExtension) (*appsv1.DaemonSet, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if !interfaces.IsNil(raw.Object) {
		if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != daemonsetsGVK {
			return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), daemonsetsGVK.String())
		}
		if obj := raw.Object.(*appsv1.DaemonSet); obj != nil {
			return obj.DeepCopy(), nil
		}
	}
	parsed := &appsv1.DaemonSet{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}

// GetContextOldDaemonSet get DaemonSet old object from Context
// no error if old object not exist, error if parse failed
func GetContextOldDaemonSet(ctx context.Context) (*appsv1.DaemonSet, error) {
	raw := util.GetContextOldObject(ctx)
	if raw == nil { // no old object
		return nil, nil
	}
	return daemonsetsRawExtensionParser(raw)
}

// deployments about

type DeploymentProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Tracer, do performance tracking
	Tracer tracer.Tracer
	// Admit do admit, return error if should stop
	Admit func(ctx context.Context, in *appsv1.Deployment) errors.APIStatus
}

type DeploymentHandler struct {
	processors []*DeploymentProcessor
	objFilters []util.ObjectIgnoreFilter
}

// deployments processor

func (p *DeploymentProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Admit == nil {
		return fmt.Errorf("%v nil processor admit function", p.Key())
	}
	return nil
}

func (p *DeploymentProcessor) DoWithTracing(ctx context.Context, in *appsv1.Deployment) (cost time.Duration, ke errors.APIStatus) {
	return p.Tracer.DoWithTracing(func() errors.APIStatus {
		return p.Admit(ctx, in)
	})
}

// deployments handler

func NewDeploymentHandler(opType arv1b1.OperationType) (review.Handler, error) {
	return handler.NewFramework(
		deploymentsGVR,
		opType,
		func(raw *runtime.RawExtension) (runtime.Object, error) {
			return deploymentsRawExtensionParser(raw)
		},
		&DeploymentHandler{},
	)
}

func (h *DeploymentHandler) IsEmpty() bool {
	return len(h.processors) == 0
}

func (h *DeploymentHandler) Register(in interface{}) error {
	getProcessor := func(v interface{}) *DeploymentProcessor {
		if v == nil {
			return nil
		}
		return v.(*DeploymentProcessor)
	}
	p := getProcessor(in)
	if p == nil {
		return errors.ErrProcessorIsNil
	}
	if e := p.Validate(); e != nil {
		return e
	}
	h.processors = append(h.processors, p)
	h.objFilters = append(h.objFilters, p.GetObjectFilter())
	return nil
}

func (h *DeploymentHandler) DoAdmit(ctx context.Context, tracer *tracer.Tracer, in runtime.Object) (cost time.Duration, ke errors.APIStatus) {
	return tracer.DoWithTracing(func() (ke errors.APIStatus) {
		// log prepare
		logBase := util.GetContextLogBase(ctx)
		// check
		obj := func() *appsv1.Deployment {
			if interfaces.IsNil(in) {
				return nil
			}
			return in.(*appsv1.Deployment)
		}()
		toFilter := obj
		if toFilter == nil {
			var err error
			toFilter, err = GetContextOldDeployment(ctx)
			if err != nil {
				err = errors.ErrWrongRuntimeObjects
				log.Errorf("%s DoAdmit failed, %v", logBase, err)
				return errors.NewBadRequest(err)
			}
		}
		// cleanup
		defer util.RemoveObjectAnno(obj, constants.AnnoKeyAdmissionIgnore)
		// execute processors
		for i, p := range h.processors {
			logPrefix := logBase + fmt.Sprintf("[%d][%s]", i, p.LogPrefix())
			// check ignore
			if ignoreReason := h.objFilters[i](toFilter); ignoreReason != nil {
				log.Infof("%s skip for %s", logPrefix, *ignoreReason)
				continue
			}
			// do review
			cost := time.Duration(0)
			select {
			case <-ctx.Done():
				ke = errors.NewRequestTimeout(errors.ErrContextEnded)
			default:
				switch p.Type {
				case constants.ProcessorTypeValidate: // do without changes
					var toValidate *appsv1.Deployment
					if obj != nil {
						toValidate = obj.DeepCopy()
					}
					cost, ke = p.DoWithTracing(ctx, toValidate)
				case constants.ProcessorTypeMutate:
					cost, ke = p.DoWithTracing(ctx, obj)
				default:
					log.Errorf("%s skip for unknown processor type '%v'", p.Type)
				}
			}
			if ke != nil {
				log.Errorf("%s[cost:%v] stop by error: %v", logPrefix, cost, ke)
				break
			}
			log.Infof("%s[cost:%v] done", logPrefix, cost)
		}
		return
	})
}

func deploymentsRawExtensionParser(raw *runtime.RawExtension) (*appsv1.Deployment, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if !interfaces.IsNil(raw.Object) {
		if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != deploymentsGVK {
			return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), deploymentsGVK.String())
		}
		if obj := raw.Object.(*appsv1.Deployment); obj != nil {
			return obj.DeepCopy(), nil
		}
	}
	parsed := &appsv1.Deployment{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}

// GetContextOldDeployment get Deployment old object from Context
// no error if old object not exist, error if parse failed
func GetContextOldDeployment(ctx context.Context) (*appsv1.Deployment, error) {
	raw := util.GetContextOldObject(ctx)
	if raw == nil { // no old object
		return nil, nil
	}
	return deploymentsRawExtensionParser(raw)
}

// replicasets about

type ReplicaSetProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Tracer, do performance tracking
	Tracer tracer.Tracer
	// Admit do admit, return error if should stop
	Admit func(ctx context.Context, in *appsv1.ReplicaSet) errors.APIStatus
}

type ReplicaSetHandler struct {
	processors []*ReplicaSetProcessor
	objFilters []util.ObjectIgnoreFilter
}

// replicasets processor

func (p *ReplicaSetProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Admit == nil {
		return fmt.Errorf("%v nil processor admit function", p.Key())
	}
	return nil
}

func (p *ReplicaSetProcessor) DoWithTracing(ctx context.Context, in *appsv1.ReplicaSet) (cost time.Duration, ke errors.APIStatus) {
	return p.Tracer.DoWithTracing(func() errors.APIStatus {
		return p.Admit(ctx, in)
	})
}

// replicasets handler

func NewReplicaSetHandler(opType arv1b1.OperationType) (review.Handler, error) {
	return handler.NewFramework(
		replicasetsGVR,
		opType,
		func(raw *runtime.RawExtension) (runtime.Object, error) {
			return replicasetsRawExtensionParser(raw)
		},
		&ReplicaSetHandler{},
	)
}

func (h *ReplicaSetHandler) IsEmpty() bool {
	return len(h.processors) == 0
}

func (h *ReplicaSetHandler) Register(in interface{}) error {
	getProcessor := func(v interface{}) *ReplicaSetProcessor {
		if v == nil {
			return nil
		}
		return v.(*ReplicaSetProcessor)
	}
	p := getProcessor(in)
	if p == nil {
		return errors.ErrProcessorIsNil
	}
	if e := p.Validate(); e != nil {
		return e
	}
	h.processors = append(h.processors, p)
	h.objFilters = append(h.objFilters, p.GetObjectFilter())
	return nil
}

func (h *ReplicaSetHandler) DoAdmit(ctx context.Context, tracer *tracer.Tracer, in runtime.Object) (cost time.Duration, ke errors.APIStatus) {
	return tracer.DoWithTracing(func() (ke errors.APIStatus) {
		// log prepare
		logBase := util.GetContextLogBase(ctx)
		// check
		obj := func() *appsv1.ReplicaSet {
			if interfaces.IsNil(in) {
				return nil
			}
			return in.(*appsv1.ReplicaSet)
		}()
		toFilter := obj
		if toFilter == nil {
			var err error
			toFilter, err = GetContextOldReplicaSet(ctx)
			if err != nil {
				err = errors.ErrWrongRuntimeObjects
				log.Errorf("%s DoAdmit failed, %v", logBase, err)
				return errors.NewBadRequest(err)
			}
		}
		// cleanup
		defer util.RemoveObjectAnno(obj, constants.AnnoKeyAdmissionIgnore)
		// execute processors
		for i, p := range h.processors {
			logPrefix := logBase + fmt.Sprintf("[%d][%s]", i, p.LogPrefix())
			// check ignore
			if ignoreReason := h.objFilters[i](toFilter); ignoreReason != nil {
				log.Infof("%s skip for %s", logPrefix, *ignoreReason)
				continue
			}
			// do review
			cost := time.Duration(0)
			select {
			case <-ctx.Done():
				ke = errors.NewRequestTimeout(errors.ErrContextEnded)
			default:
				switch p.Type {
				case constants.ProcessorTypeValidate: // do without changes
					var toValidate *appsv1.ReplicaSet
					if obj != nil {
						toValidate = obj.DeepCopy()
					}
					cost, ke = p.DoWithTracing(ctx, toValidate)
				case constants.ProcessorTypeMutate:
					cost, ke = p.DoWithTracing(ctx, obj)
				default:
					log.Errorf("%s skip for unknown processor type '%v'", p.Type)
				}
			}
			if ke != nil {
				log.Errorf("%s[cost:%v] stop by error: %v", logPrefix, cost, ke)
				break
			}
			log.Infof("%s[cost:%v] done", logPrefix, cost)
		}
		return
	})
}

func replicasetsRawExtensionParser(raw *runtime.RawExtension) (*appsv1.ReplicaSet, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if !interfaces.IsNil(raw.Object) {
		if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != replicasetsGVK {
			return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), replicasetsGVK.String())
		}
		if obj := raw.Object.(*appsv1.ReplicaSet); obj != nil {
			return obj.DeepCopy(), nil
		}
	}
	parsed := &appsv1.ReplicaSet{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}

// GetContextOldReplicaSet get ReplicaSet old object from Context
// no error if old object not exist, error if parse failed
func GetContextOldReplicaSet(ctx context.Context) (*appsv1.ReplicaSet, error) {
	raw := util.GetContextOldObject(ctx)
	if raw == nil { // no old object
		return nil, nil
	}
	return replicasetsRawExtensionParser(raw)
}

// statefulsets about

type StatefulSetProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Tracer, do performance tracking
	Tracer tracer.Tracer
	// Admit do admit, return error if should stop
	Admit func(ctx context.Context, in *appsv1.StatefulSet) errors.APIStatus
}

type StatefulSetHandler struct {
	processors []*StatefulSetProcessor
	objFilters []util.ObjectIgnoreFilter
}

// statefulsets processor

func (p *StatefulSetProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Admit == nil {
		return fmt.Errorf("%v nil processor admit function", p.Key())
	}
	return nil
}

func (p *StatefulSetProcessor) DoWithTracing(ctx context.Context, in *appsv1.StatefulSet) (cost time.Duration, ke errors.APIStatus) {
	return p.Tracer.DoWithTracing(func() errors.APIStatus {
		return p.Admit(ctx, in)
	})
}

// statefulsets handler

func NewStatefulSetHandler(opType arv1b1.OperationType) (review.Handler, error) {
	return handler.NewFramework(
		statefulsetsGVR,
		opType,
		func(raw *runtime.RawExtension) (runtime.Object, error) {
			return statefulsetsRawExtensionParser(raw)
		},
		&StatefulSetHandler{},
	)
}

func (h *StatefulSetHandler) IsEmpty() bool {
	return len(h.processors) == 0
}

func (h *StatefulSetHandler) Register(in interface{}) error {
	getProcessor := func(v interface{}) *StatefulSetProcessor {
		if v == nil {
			return nil
		}
		return v.(*StatefulSetProcessor)
	}
	p := getProcessor(in)
	if p == nil {
		return errors.ErrProcessorIsNil
	}
	if e := p.Validate(); e != nil {
		return e
	}
	h.processors = append(h.processors, p)
	h.objFilters = append(h.objFilters, p.GetObjectFilter())
	return nil
}

func (h *StatefulSetHandler) DoAdmit(ctx context.Context, tracer *tracer.Tracer, in runtime.Object) (cost time.Duration, ke errors.APIStatus) {
	return tracer.DoWithTracing(func() (ke errors.APIStatus) {
		// log prepare
		logBase := util.GetContextLogBase(ctx)
		// check
		obj := func() *appsv1.StatefulSet {
			if interfaces.IsNil(in) {
				return nil
			}
			return in.(*appsv1.StatefulSet)
		}()
		toFilter := obj
		if toFilter == nil {
			var err error
			toFilter, err = GetContextOldStatefulSet(ctx)
			if err != nil {
				err = errors.ErrWrongRuntimeObjects
				log.Errorf("%s DoAdmit failed, %v", logBase, err)
				return errors.NewBadRequest(err)
			}
		}
		// cleanup
		defer util.RemoveObjectAnno(obj, constants.AnnoKeyAdmissionIgnore)
		// execute processors
		for i, p := range h.processors {
			logPrefix := logBase + fmt.Sprintf("[%d][%s]", i, p.LogPrefix())
			// check ignore
			if ignoreReason := h.objFilters[i](toFilter); ignoreReason != nil {
				log.Infof("%s skip for %s", logPrefix, *ignoreReason)
				continue
			}
			// do review
			cost := time.Duration(0)
			select {
			case <-ctx.Done():
				ke = errors.NewRequestTimeout(errors.ErrContextEnded)
			default:
				switch p.Type {
				case constants.ProcessorTypeValidate: // do without changes
					var toValidate *appsv1.StatefulSet
					if obj != nil {
						toValidate = obj.DeepCopy()
					}
					cost, ke = p.DoWithTracing(ctx, toValidate)
				case constants.ProcessorTypeMutate:
					cost, ke = p.DoWithTracing(ctx, obj)
				default:
					log.Errorf("%s skip for unknown processor type '%v'", p.Type)
				}
			}
			if ke != nil {
				log.Errorf("%s[cost:%v] stop by error: %v", logPrefix, cost, ke)
				break
			}
			log.Infof("%s[cost:%v] done", logPrefix, cost)
		}
		return
	})
}

func statefulsetsRawExtensionParser(raw *runtime.RawExtension) (*appsv1.StatefulSet, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if !interfaces.IsNil(raw.Object) {
		if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != statefulsetsGVK {
			return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), statefulsetsGVK.String())
		}
		if obj := raw.Object.(*appsv1.StatefulSet); obj != nil {
			return obj.DeepCopy(), nil
		}
	}
	parsed := &appsv1.StatefulSet{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}

// GetContextOldStatefulSet get StatefulSet old object from Context
// no error if old object not exist, error if parse failed
func GetContextOldStatefulSet(ctx context.Context) (*appsv1.StatefulSet, error) {
	raw := util.GetContextOldObject(ctx)
	if raw == nil { // no old object
		return nil, nil
	}
	return statefulsetsRawExtensionParser(raw)
}

// workloads about

type WorkloadProcessor struct {
	// Metadata, set name, type and ignore settings
	processor.Metadata
	// Tracer, do performance tracking
	Tracer tracer.Tracer
	// Admit do admit, return error if should stop
	Admit func(ctx context.Context, in *wklv1a1.Workload) errors.APIStatus
}

type WorkloadHandler struct {
	processors []*WorkloadProcessor
	objFilters []util.ObjectIgnoreFilter
}

// workloads processor

func (p *WorkloadProcessor) Validate() error {
	if e := p.Metadata.Validate(); e != nil {
		return e
	}
	if p.Admit == nil {
		return fmt.Errorf("%v nil processor admit function", p.Key())
	}
	return nil
}

func (p *WorkloadProcessor) DoWithTracing(ctx context.Context, in *wklv1a1.Workload) (cost time.Duration, ke errors.APIStatus) {
	return p.Tracer.DoWithTracing(func() errors.APIStatus {
		return p.Admit(ctx, in)
	})
}

// workloads handler

func NewWorkloadHandler(opType arv1b1.OperationType) (review.Handler, error) {
	return handler.NewFramework(
		workloadsGVR,
		opType,
		func(raw *runtime.RawExtension) (runtime.Object, error) {
			return workloadsRawExtensionParser(raw)
		},
		&WorkloadHandler{},
	)
}

func (h *WorkloadHandler) IsEmpty() bool {
	return len(h.processors) == 0
}

func (h *WorkloadHandler) Register(in interface{}) error {
	getProcessor := func(v interface{}) *WorkloadProcessor {
		if v == nil {
			return nil
		}
		return v.(*WorkloadProcessor)
	}
	p := getProcessor(in)
	if p == nil {
		return errors.ErrProcessorIsNil
	}
	if e := p.Validate(); e != nil {
		return e
	}
	h.processors = append(h.processors, p)
	h.objFilters = append(h.objFilters, p.GetObjectFilter())
	return nil
}

func (h *WorkloadHandler) DoAdmit(ctx context.Context, tracer *tracer.Tracer, in runtime.Object) (cost time.Duration, ke errors.APIStatus) {
	return tracer.DoWithTracing(func() (ke errors.APIStatus) {
		// log prepare
		logBase := util.GetContextLogBase(ctx)
		// check
		obj := func() *wklv1a1.Workload {
			if interfaces.IsNil(in) {
				return nil
			}
			return in.(*wklv1a1.Workload)
		}()
		toFilter := obj
		if toFilter == nil {
			var err error
			toFilter, err = GetContextOldWorkload(ctx)
			if err != nil {
				err = errors.ErrWrongRuntimeObjects
				log.Errorf("%s DoAdmit failed, %v", logBase, err)
				return errors.NewBadRequest(err)
			}
		}
		// cleanup
		defer util.RemoveObjectAnno(obj, constants.AnnoKeyAdmissionIgnore)
		// execute processors
		for i, p := range h.processors {
			logPrefix := logBase + fmt.Sprintf("[%d][%s]", i, p.LogPrefix())
			// check ignore
			if ignoreReason := h.objFilters[i](toFilter); ignoreReason != nil {
				log.Infof("%s skip for %s", logPrefix, *ignoreReason)
				continue
			}
			// do review
			cost := time.Duration(0)
			select {
			case <-ctx.Done():
				ke = errors.NewRequestTimeout(errors.ErrContextEnded)
			default:
				switch p.Type {
				case constants.ProcessorTypeValidate: // do without changes
					var toValidate *wklv1a1.Workload
					if obj != nil {
						toValidate = obj.DeepCopy()
					}
					cost, ke = p.DoWithTracing(ctx, toValidate)
				case constants.ProcessorTypeMutate:
					cost, ke = p.DoWithTracing(ctx, obj)
				default:
					log.Errorf("%s skip for unknown processor type '%v'", p.Type)
				}
			}
			if ke != nil {
				log.Errorf("%s[cost:%v] stop by error: %v", logPrefix, cost, ke)
				break
			}
			log.Infof("%s[cost:%v] done", logPrefix, cost)
		}
		return
	})
}

func workloadsRawExtensionParser(raw *runtime.RawExtension) (*wklv1a1.Workload, error) {
	if raw == nil {
		return nil, fmt.Errorf("runtime.RawExtension is nil")
	}
	if !interfaces.IsNil(raw.Object) {
		if gvk := raw.Object.GetObjectKind().GroupVersionKind(); gvk != workloadsGVK {
			return nil, fmt.Errorf("runtime.RawExtension group version kind '%v' != '%v'", gvk.String(), workloadsGVK.String())
		}
		if obj := raw.Object.(*wklv1a1.Workload); obj != nil {
			return obj.DeepCopy(), nil
		}
	}
	parsed := &wklv1a1.Workload{}
	if e := json.Unmarshal(raw.Raw, parsed); e != nil {
		return nil, e
	}
	return parsed, nil
}

// GetContextOldWorkload get Workload old object from Context
// no error if old object not exist, error if parse failed
func GetContextOldWorkload(ctx context.Context) (*wklv1a1.Workload, error) {
	raw := util.GetContextOldObject(ctx)
	if raw == nil { // no old object
		return nil, nil
	}
	return workloadsRawExtensionParser(raw)
}
