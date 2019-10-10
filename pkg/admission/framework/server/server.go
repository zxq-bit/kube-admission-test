package server

import (
	"time"

	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/module"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/review/manager"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/server/initialize"

	"github.com/caicloud/clientset/informers"
	"github.com/caicloud/clientset/kubernetes"
	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/config"
	"github.com/caicloud/nirvana/log"

	"k8s.io/client-go/tools/clientcmd"
)

type Server struct {
	// cmd & parse
	cfg initialize.Config
	cmd config.NirvanaCommand
	// parsed
	enableOptions []string
	caBundle      []byte
	certFile      string
	keyFile       string
	handlerConfig *initialize.HandlerConfig

	// kube
	kc              kubernetes.Interface
	informerFactory informers.SharedInformerFactory

	// module & review
	moduleManager module.Manager
	reviewManager manager.Manager

	stopCh chan struct{}
}

func NewServer() (*Server, error) {
	s := &Server{
		cfg: initialize.DefaultServerConfig(),
		cmd: config.NewNirvanaCommand(&config.Option{
			Port: uint16(constants.DefaultListenPort),
		}),
		stopCh: make(chan struct{}),
	}
	s.cmd.AddOption("", &s.cfg)
	s.cmd.SetHook(&config.NirvanaCommandHookFunc{
		PostConfigureFunc: s.init,
		PostServeFunc:     s.postServe,
	})
	return s, nil
}

func (s *Server) Run() error {
	return s.cmd.Execute()
}

func (s *Server) init(config *nirvana.Config) error {
	kubeHost := s.cfg.KubeHost
	kubeConfig := s.cfg.KubeConfig
	log.Infof("parsed config: %s", s.cfg.String())
	// config
	if e := s.validateConfig(); e != nil {
		log.Errorf("validateConfig failed, %v", e)
		return e
	}
	log.Infof("validateConfig done")
	// kube
	restConf, e := clientcmd.BuildConfigFromFlags(kubeHost, kubeConfig)
	if e != nil {
		log.Errorf("BuildConfigFromFlags failed, %v", e)
		return e
	}
	log.Infof("BuildConfigFromFlags done")
	s.kc, e = kubernetes.NewForConfig(restConf)
	if e != nil {
		log.Errorf("NewForConfig failed, %v", e)
		return e
	}
	log.Infof("NewForConfig done")
	s.informerFactory = informers.NewSharedInformerFactory(s.kc,
		time.Duration(s.cfg.InformerFactoryResyncSecond)*time.Second)
	// cert
	if e = s.ensureCert(); e != nil {
		log.Errorf("ensureCert failed, %v", e)
		return e
	}
	log.Infof("ensureCert done")

	// init
	// init selected modules
	s.ensureModuleMaker()
	if e = s.initModules(); e != nil {
		log.Errorf("initModules failed, %v", e)
		return e
	}
	log.Infof("initModules done")
	// init selected processors
	if e = s.initReviews(); e != nil {
		log.Errorf("initReviews failed, %v", e)
		return e
	}
	log.Infof("initReviews done")

	// start
	// modules
	if e = s.startModules(); e != nil {
		log.Errorf("startModules failed, %v", e)
		return e
	}
	log.Infof("startModules done")
	// nirvana
	log.Infof("s.cfg.certFile:%s", s.certFile)
	log.Infof("s.cfg.keyFile:%s", s.keyFile)
	config.Configure(
		nirvana.Descriptor(s.reviewManager.GetDescriptors()...),
		nirvana.TLS(s.certFile, s.keyFile),
	)
	log.Infof("Configure done")
	// webhooks
	if e = s.ensureWebhooks(); e != nil {
		log.Errorf("ensureWebhooks failed, %v", e)
		return e
	}
	log.Infof("ensureWebhooks done")

	return nil
}

func (s *Server) postServe(_ *nirvana.Config, _ nirvana.Server, _ error) error {
	close(s.stopCh)
	return nil
}
