package naming

import (
	"github.com/inooy/serco-client/core"
	"github.com/inooy/serco-client/pkg/tools"
)

type ServiceManager struct {
	AppId      string
	Env        string
	InstanceId string
	Client     *core.SocketClientImpl
}

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
}

func NewNamingService(appId string, env string, client *core.SocketClientImpl) *ServiceManager {
	manager := ServiceManager{
		Client:     client,
		AppId:      appId,
		Env:        env,
		InstanceId: appId + tools.GetSnowflakeId(),
	}
	manager.Client.OnReconnected(func(isReconnect bool) error {
		// 重连，需要重新注册
		return manager.Registry()
	})
	return &manager
}

func Setup(autoRegistry bool) {

}
