package serco

import (
	"github.com/inooy/serco-client/config"
	"github.com/inooy/serco-client/core"
	"github.com/inooy/serco-client/naming"
	"github.com/inooy/serco-client/pkg/log"
	"github.com/inooy/serco-client/pkg/tools"
	"sync"
)

// Options config base options
type Options struct {
	// config env
	EnvId string
	// the appName of at config center
	AppName string
	// Configure the center addresses. Multiple addresses are separated by commas(,)
	RemoteAddr string
	// Configure polling interval in milliseconds.
	// This mechanism mainly avoids the loss of change notice
	PollInterval int
	InstanceId   string
}

type RegistryOpts struct {
	LocalIp   string
	LocalPort int
	Protocol  string
}

type Serco struct {
	once           sync.Once
	Options        *Options
	Client         *core.SocketClientImpl
	ConfigManager  *config.Manager
	ServiceManager *naming.ServiceManager
}

func NewSerco(options Options) *Serco {
	if options.InstanceId == "" {
		options.InstanceId = options.AppName + tools.GetSnowflakeId()
	}
	instance := &Serco{
		Options: &options,
	}

	if options.RemoteAddr != "" {
		instance.Client = core.NewSocketClient(&core.Options{
			EnvId:      options.EnvId,
			AppName:    options.AppName,
			RemoteAddr: options.RemoteAddr,
		})

		err := instance.Client.Launch()
		if err != nil {
			panic(err)
		}
	}

	return instance
}

func (s *Serco) SetupConfig(bean interface{}) {
	s.ConfigManager = config.NewManager(&config.Options{
		EnvId:        s.Options.EnvId,
		AppName:      s.Options.AppName,
		RemoteAddr:   s.Options.RemoteAddr,
		PollInterval: s.Options.PollInterval,
	}, bean, s.Client)
	s.ConfigManager.InitConfig()
}

func (s *Serco) SetupDiscovery(opt RegistryOpts) {
	if opt.Protocol == "" {
		opt.Protocol = "http"
	}
	s.ServiceManager = naming.NewNamingService(&naming.Options{
		EnvId:        s.Options.EnvId,
		AppName:      s.Options.AppName,
		RemoteAddr:   s.Options.RemoteAddr,
		PollInterval: s.Options.PollInterval,
		InstanceId:   s.Options.InstanceId,
		LocalIp:      opt.LocalIp,
		LocalPort:    opt.LocalPort,
		Protocol:     opt.Protocol,
	}, s.Client)
}

func (s *Serco) Registry() error {
	if s.Client == nil {
		log.Warn("serco client not provide, skip register")
		return nil
	}
	return s.ServiceManager.Registry()
}

func (s *Serco) GetInstance(appName string) ([]*naming.Instance, error) {
	return s.ServiceManager.GetInstance(appName)
}

func (s *Serco) Subscribe(providers []*naming.SubscribeProvider) error {
	if s.Client == nil {
		log.Warn("serco client not provide, skip subscribe")
		return nil
	}
	return s.ServiceManager.Subscribe(providers)
}

func (s *Serco) Shutdown() error {
	if s.ConfigManager != nil {
		s.ConfigManager.Shutdown()
	}
	if s.ServiceManager != nil {
		err := s.ServiceManager.Shutdown()
		if err != nil {
			return err
		}
	}
	return nil
}
