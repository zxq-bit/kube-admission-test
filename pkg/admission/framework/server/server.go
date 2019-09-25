package server

import (
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/configs"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/constants"
	"github.com/zxq-bit/kube-admission-test/pkg/admission/framework/model"

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

	modelCollection           model.Collection
	processorConfigCollection configs.Collection

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
	e := s.cfg.Validate()
	if e != nil {
		log.Errorf("validate config failed, %v", e)
		return e
	}
	log.Info(s.cfg.String())
	// kube
	restConf, e := clientcmd.BuildConfigFromFlags(kubeHost, kubeConfig)
	if e != nil {
		log.Errorf("BuildConfigFromFlags failed, %v", e)
		return e
	}
	s.kc, e = kubernetes.NewForConfig(restConf)
	if e != nil {
		return e
	}
	s.informerFactory = informers.NewSharedInformerFactory(s.kc, s.cfg.InformerFactoryResync)
	// cert
	caBundle, certFile, keyFile, e := s.ensureCert()
	if e != nil {
		return e
	}
	opt := s.cfg.ToStartOptions()
	opt.ServiceCABundle = caBundle

	// init
	if e = s.initModelsAndProcessors(); e != nil {
		log.Errorf("initModelsAndProcessors failed, %v", e)
		return e
	}

	// start
	if e = s.startModels(opt); e != nil {
		log.Errorf("startModels failed, %v", e)
		return e
	}
	config.Configure(
		nirvana.Descriptor(s.processorConfigCollection.GetDescriptors(&opt)...),
		nirvana.TLS(certFile, keyFile),
	)

	return nil
}

func (s *Server) postServe(_ *nirvana.Config, _ nirvana.Server, _ error) error {
	close(s.stopCh)
	return nil
}
