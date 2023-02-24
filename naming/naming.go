package naming

import (
	"github.com/inooy/serco-client/core"
	"github.com/inooy/serco-client/pkg/log"
)

type ServiceManager struct {
	Options    *Options
	Registered bool
	Client     *core.SocketClientImpl
	App        map[string]*Instance
	Providers  []*SubscribeProvider
}

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
	LocalIp      string
	LocalPort    int
	Protocol     string
}

func NewNamingService(options *Options, client *core.SocketClientImpl) *ServiceManager {
	manager := ServiceManager{
		Client:    client,
		Options:   options,
		Providers: make([]*SubscribeProvider, 0),
	}
	if client == nil {
		return &manager
	}
	manager.Client.OnReconnected(func(isReconnect bool) error {
		var err error
		if manager.Registered {
			// 重连，需要重新注册
			err = manager.Registry()
		} else {
			log.Info("service not registry, reconnect skip registry")
		}
		if err == nil && len(manager.Providers) > 0 {
			// 重新订阅
			err = manager.Subscribe(manager.Providers)
		}
		return err
	})
	return &manager
}

func (m *ServiceManager) GetInstance(appName string) ([]*Instance, error) {

	return nil, nil
}

func (m *ServiceManager) Subscribe(providers []*SubscribeProvider) error {
	log.Info("start subscribe service")
	if m.Client == nil {
		return nil
	}
	subscribe := SubscribeCmd{
		InstanceId: m.Options.InstanceId,
		Subscribes: providers,
	}

	for i := range providers {
		m.Providers = append(m.Providers, providers[i])
	}

	// 订阅服务
	return m.SubscribeRequest(subscribe)
}

func (m *ServiceManager) Shutdown() error {
	return m.Cancel()
}

func (m *ServiceManager) poll() {
	// 定期拉取服务列表
}
