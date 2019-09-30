package initialize

import (
	"encoding/json"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
)

type Config struct {
	// kube
	KubeHost   string `desc:"kubernetes host"`
	KubeConfig string `desc:"kubernetes config"`
	// informer
	InformerFactoryResyncSecond int32 `desc:"kubernetes informer factory resync time by seconds"`
	// admit
	ServiceNamespace string `desc:"admission service namespace"`
	ServiceName      string `desc:"admission service name"`
	ServiceSelector  string `desc:"admission service selector labels key value pairs"`
	CertTempDir      string `desc:"admission server cert file template dir path"`
	// enable
	Admissions string `desc:"a list of admissions to enable. '*' enables all on-by-default admissions"`
	// review config
	ReviewConfigFile string `desc:"review config file path"`
}

func DefaultServerConfig() Config {
	return Config{
		InformerFactoryResyncSecond: constants.DefaultInformerFactoryResyncSecond,
		ServiceNamespace:            constants.DefaultServiceNamespace,
		ServiceName:                 constants.DefaultServiceName,
		ServiceSelector:             constants.DefaultServiceSelector,
		CertTempDir:                 constants.DefaultCertTempDir,
		Admissions:                  constants.AdmissionsAll,
		ReviewConfigFile:            constants.DefaultReviewConfigFilePath,
	}
}

func (c *Config) String() string {
	b, _ := json.MarshalIndent(c, "", "  ")
	return string(b)
}
