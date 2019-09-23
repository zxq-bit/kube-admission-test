package resources

import (
	"reflect"
	"sort"

	corev1 "k8s.io/api/core/v1"

	"github.com/caicloud/workload/pkg/type/v1alpha1"
)

// MutateTools common mutate tools
type MutateTools struct {
	Namespace   string
	ReleaseName string
	Kind        string
}

var defaultEnv = map[string]corev1.EnvVar{
	"POD_NAMESPACE": {
		Name: "POD_NAMESPACE",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.namespace",
			},
		},
	},
	"POD_NAME": {
		Name: "POD_NAME",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.name",
			},
		},
	},
	"POD_IP": {
		Name: "POD_IP",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "status.podIP",
			},
		},
	},
	"NODE_NAME": {
		Name: "NODE_NAME",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "spec.nodeName",
			},
		},
	},
}

func (mt *MutateTools) defaultAnnotations() map[string]string {
	return map[string]string{
		v1alpha1.WorkloadAnnotationCreateBy: "yaml",
	}
}

// forceAnnotations the annotation keep must key same with this
func (mt *MutateTools) forceAnnotations() map[string]string {
	return map[string]string{
		// use for event and monitoring data
		v1alpha1.HelmAnnotationName:      mt.ReleaseName,
		v1alpha1.HelmAnnotationNamespace: mt.Namespace,
		v1alpha1.HelmAnnotationPath:      "app",
	}
}

func (mt *MutateTools) defaultLabels() map[string]string {
	m := make(map[string]string)
	m[v1alpha1.WorkloadTaintLabelsName] = mt.ReleaseName
	m[v1alpha1.WorkloadTaintLabelsKind] = mt.Kind
	return m
}

func (mt *MutateTools) defaultMatchLabels() map[string]string {
	m := make(map[string]string)
	m[v1alpha1.WorkloadTaintLabelsName] = mt.ReleaseName
	m[v1alpha1.WorkloadTaintLabelsKind] = mt.Kind
	return m
	// TODO: why not merge labels and matchLabels?
}

// overwriteMaps merges source map to target map and returns target map force
func overwriteMaps(source, target map[string]string) map[string]string {
	if target == nil {
		target = make(map[string]string)
	}
	// add keys in source to target
	for k, v := range source {
		target[k] = v
	}
	return target
}

// mergeMaps merges source map to target map and returns target map
// will keep the value exists in target
func mergeMaps(source, target map[string]string) map[string]string {
	if target == nil {
		target = make(map[string]string)
	}
	// add keys in source to target
	for k, v := range source {
		if _, exist := target[k]; !exist {
			target[k] = v
		}
	}
	return target
}

// MutateAnnotations mutates annotations
func MutateAnnotations(mt *MutateTools, anno map[string]string) map[string]string {
	m := mergeMaps(mt.defaultAnnotations(), anno)
	return overwriteMaps(mt.forceAnnotations(), m)
}

// MutateLabels mutates labels
func MutateLabels(mt *MutateTools, l map[string]string) map[string]string {
	return overwriteMaps(mt.defaultLabels(), l)
}

// MutateMatchLabels mutates match labels
func MutateMatchLabels(mt *MutateTools, ml map[string]string) map[string]string {
	return overwriteMaps(mt.defaultMatchLabels(), ml)
}

// MutateContainers mutates containers
func MutateContainers(mt *MutateTools, containers []corev1.Container) []corev1.Container {
	for i := range containers {
		env := make(map[string]int)
		container := &containers[i]
		// record env names and their indices for later modification
		for j, v := range container.Env {
			env[v.Name] = j
		}
		for key, v := range defaultEnv {
			j, ok := env[key]
			// if some key in default env doesn't exist in container's env, append it
			if !ok {
				container.Env = append(container.Env, v)
				continue
			}
			// if some key exists but values differ in default and container, update container
			if !reflect.DeepEqual(v, container.Env[j]) {
				container.Env[j] = v
			}
		}
		// make sure env is ordered, otherwise will lead workload to increase version
		sort.Slice(containers[i].Env, func(i1, i2 int) bool {
			return containers[i].Env[i1].Name < containers[i].Env[i2].Name
		})
	}
	// sort containers based on their names
	sort.Slice(containers, func(i, j int) bool {
		return containers[i].Name < containers[j].Name
	})
	return containers
}

// MutateTemplate mutates template
func MutateTemplate(mt *MutateTools, template *corev1.PodTemplateSpec) {
	// ensure merged labels into template labels and selector match labels are the same
	template.Labels = MutateMatchLabels(mt, template.Labels)
	template.Annotations = MutateAnnotations(mt, template.Annotations)
	template.Spec.Containers = MutateContainers(mt, template.Spec.Containers)
	template.Spec.InitContainers = MutateContainers(mt, template.Spec.InitContainers)
}
