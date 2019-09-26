package app

const (
	ModelName = "app"

	ProcessorNamePodGPUVisible = "PodGPUVisible"
)

const (
	envKeyNvidiaVisibleDevices = "NVIDIA_VISIBLE_DEVICES"
	resourceKeyNvGPU           = "nvidia.com/gpu"
	resourceKeyPrefixER        = "extendedresource.caicloud.io/"
	resourceKeyPrefixERReq     = "requests.extendedresource.caicloud.io/"
)
