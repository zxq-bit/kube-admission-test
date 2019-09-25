package server

import (
	"io/ioutil"
	"path/filepath"

	"github.com/caicloud/go-common/cert"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
)

func (s *Server) ensureCert() (caBundle []byte, certFile, keyFile string, err error) {
	certData, keyData, err := cert.GenSelfSignedCertForK8sService(s.cfg.ServiceNamespace, s.cfg.ServiceName)
	if err != nil {
		return
	}
	caBundle = certData
	certFile = filepath.Join(s.cfg.CertTempDir, constants.DefaultCertFileName)
	keyFile = filepath.Join(s.cfg.CertTempDir, constants.DefaultKeyFileName)
	dataPath := map[string][]byte{
		certFile: certData,
		keyFile:  keyData,
	}
	for fp, data := range dataPath {
		if err = ioutil.WriteFile(fp, data, 0664); err != nil {
			return
		}
	}
	return
}
