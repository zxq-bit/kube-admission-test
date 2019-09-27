package server

import (
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/interfaces"

	"github.com/caicloud/clientset/informers"
	"github.com/caicloud/clientset/kubernetes"
	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/config"
	"github.com/caicloud/nirvana/log"

	"k8s.io/client-go/tools/clientcmd"
)

type Server struct {
	cfg Config
	cmd config.NirvanaCommand

	kc              kubernetes.Interface
	informerFactory informers.SharedInformerFactory

	modelCollection  interfaces.ModelCollection
	configCollection interfaces.ConfigCollection

	stopCh chan struct{}
}

func NewServer() (*Server, error) {
	s := &Server{
		cfg: *NewDefaultConfig(),
		cmd: config.NewNirvanaCommand(&config.Option{
			Port: uint16(constants.DefaultListenPort),
		}),
		stopCh: make(chan struct{}),
	}
	s.cmd.AddOption("", &s.cfg)
	s.cmd.SetHook(&config.NirvanaCommandHookFunc{
		PreConfigureFunc: s.init,
		PostServeFunc:    s.postServe,
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
	if e := s.cfg.Validate(); e != nil {
		log.Errorf("validate config failed, %v", e)
		return e
	}
	log.Infof("Validate done")
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
	s.informerFactory = informers.NewSharedInformerFactory(s.kc, s.cfg.InformerFactoryResync)
	// cert
	caBundle, certFile, keyFile, e := s.ensureCert()
	if e != nil {
		log.Errorf("ensureCert failed, %v", e)
		return e
	}
	log.Infof("ensureCert done")
	opt := s.cfg.ToStartOptions()
	opt.ServiceCABundle = caBundle

	// init
	if e = s.initModelsAndProcessors(); e != nil {
		log.Errorf("initModelsAndProcessors failed, %v", e)
		return e
	}
	log.Infof("initModelsAndProcessors done")

	// start
	if e = s.startModels(); e != nil {
		log.Errorf("startModels failed, %v", e)
		return e
	}
	log.Infof("startModels done")
	config.Configure(
		nirvana.Descriptor(s.configCollection.GetDescriptors(&opt)...),
		nirvana.TLS(certFile, keyFile),
	)
	log.Infof("Configure done")

	// service
	if e = s.ensureService(int(config.Port())); e != nil {
		log.Errorf("ensureService failed, %v", e)
		return e
	}
	log.Infof("ensureService done")

	// webhooks
	if e = s.ensureWebhooks(&opt); e != nil {
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
