package serco

import (
	"github.com/inooy/serco-client/config"
	"github.com/inooy/serco-client/core"
	"github.com/inooy/serco-client/pkg/socket/connection"
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
}

type Serco struct {
	once          sync.Once
	Options       *Options
	Client        *core.SocketClientImpl
	configManager *config.Manager
}

func NewSerco(options Options) *Serco {
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

func (s *Serco) Registry() {

}

func (s *Serco) Shutdown() {
	if s.configManager != nil {
		s.configManager.Shutdown()
	}

}
