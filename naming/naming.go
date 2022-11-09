package naming

import (
	"github.com/inooy/serco-client/config/remote"
	"github.com/inooy/serco-client/pkg/tools"
)

type ServiceManager struct {
	AppId      string
	Env        string
	InstanceId string
	Client     *remote.SercoSocketClient
}

func NewNamingService(appId string, env string, client *remote.SercoSocketClient) *ServiceManager {
	manager := ServiceManager{
		Client:     client,
		AppId:      appId,
		Env:        env,
		InstanceId: appId + tools.GetSnowflakeId(),
	}
	return &manager
}
