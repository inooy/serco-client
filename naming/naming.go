package naming

import (
	"github.com/inooy/serco-client/core"
)

type ServiceManager struct {
	Options   *Options
	Client    *core.SocketClientImpl
	App       map[string]*Instance
	Providers []*SubscribeProvider
}

// Options config base options
type Options struct {
	// config env
	EnvType string
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
		Client:  client,
		Options: options,
	}
	manager.Client.OnReconnected(func(isReconnect bool) error {
		// 重连，需要重新注册
		return manager.Registry()
	})
	return &manager
}

func (m *ServiceManager) GetInstance(appName string) ([]*Instance, error) {

	return nil, nil
}

func (m *ServiceManager) Subscribe(providers []*SubscribeProvider) error {
	subscribe := SubscribeCmd{
		InstanceId: m.Options.InstanceId,
		Subscribes: providers,
	}

	// 订阅服务
	return m.SubscribeRequest(subscribe)
}

func (m *ServiceManager) Shutdown() error {
	return m.Cancel()
}

func (m *ServiceManager) poll() {

}
