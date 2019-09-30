package demo

const (
	ModelName = "demo"

	ProcessorNameCmExample     = "ConfigMapExample"
	ProcessorNamePodExample    = "PodExample"
	ProcessorNamePodGPUVisible = "PodGPUVisible"
)

const (
	envKeyNvidiaVisibleDevices = "NVIDIA_VISIBLE_DEVICES"
	resourceKeyNvGPU           = "nvidia.com/gpu"
	resourceKeyPrefixER        = "extendedresource.caicloud.io/"
	resourceKeyPrefixERReq     = "requests.extendedresource.caicloud.io/"
)
