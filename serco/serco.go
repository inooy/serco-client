package serco

import (
	"github.com/inooy/serco-client/config"
	"github.com/inooy/serco-client/core"
	"github.com/inooy/serco-client/naming"
	"github.com/inooy/serco-client/pkg/socket/connection"
	"github.com/inooy/serco-client/pkg/tools"
	"sync"
)

// Options config base options
type Options struct {
	// config env
	Env string
	// the appName of at config center
	AppName string
	// Configure the center addresses. Multiple addresses are separated by commas(,)
	RemoteAddr string
	// Configure polling interval in milliseconds.
	// This mechanism mainly avoids the loss of change notice
	PollInterval int
	// use registry center
	RegistryEnabled bool
	// use config center
	ConfigEnabled bool
	InstanceId    string
}

type Serco struct {
	once           sync.Once
	Options        *Options
	Client         *core.SocketClientImpl
	configManager  *config.Manager
	serviceManager *naming.ServiceManager
}

func NewSerco(options Options) *Serco {
	if options.InstanceId == "" {
		options.InstanceId = options.AppName + tools.GetSnowflakeId()
	}
	instance := &Serco{
		Options: &options,
	}

	instance.Client = core.NewConfigSocketClient(connection.TcpSocketConnectOpts{
		Host: options.RemoteAddr,
	})

	return instance
}

func (s *Serco) SetupConfig(bean interface{}) {
	s.configManager = config.NewManager(&config.Options{
		Env:          s.Options.Env,
		AppName:      s.Options.AppName,
		RemoteAddr:   s.Options.RemoteAddr,
		PollInterval: s.Options.PollInterval,
	}, bean, s.Client)
	s.configManager.InitConfig()
}

func (s *Serco) Registry() error {
	s.serviceManager = naming.NewNamingService(&naming.Options{
		Env:          s.Options.Env,
		AppName:      s.Options.AppName,
		RemoteAddr:   s.Options.RemoteAddr,
		PollInterval: s.Options.PollInterval,
		InstanceId:   s.Options.InstanceId,
	}, s.Client)
	return s.serviceManager.Registry()
}

func (s *Serco) GetInstance(appName string) ([]*naming.Instance, error) {
	return s.serviceManager.GetInstance(appName)
}

func (s *Serco) Subscribe(providers []*naming.SubscribeProvider) error {
	return s.serviceManager.Subscribe(providers)
}

func (s *Serco) Shutdown() error {
	if s.configManager != nil {
		s.configManager.Shutdown()
	}
	if s.serviceManager != nil {
		err := s.serviceManager.Shutdown()
		if err != nil {
			return err
		}
	}
	return nil
}
