package naming

// RegisterCmd 服务注册参数
type RegisterCmd struct {
	EnvId       string            `json:"envId" mapstructure:"envId"`
	AppId       string            `json:"appId" mapstructure:"appId"`
	InstanceId  string            `json:"instanceId" mapstructure:"instanceId"`
	AddressList []string          `json:"addressList" mapstructure:"addressList"` // 服务实例的地址，可以是 http 或 rpc 地址，多个地址可以维护数组
	Status      uint32            `json:"status" mapstructure:"status"`
	Version     string            `json:"version" mapstructure:"version"`
	LatestTime  int64             `json:"latestTime" mapstructure:"latestTime"`
	DirtyTime   int64             `json:"dirtyTime" mapstructure:"dirtyTime"`     //other node send
	Replication bool              `json:"replication" mapstructure:"replication"` //other node send
	Metadata    map[string]string `json:"metadata" mapstructure:"metadata"`
}

// SubscribeCmd 服务订阅参数
type SubscribeCmd struct {
	InstanceId string               `json:"instanceId" mapstructure:"instanceId"`
	Subscribes []*SubscribeProvider `json:"subscribes" mapstructure:"subscribes"`
}

type SubscribeProvider struct {
	Provider string            `json:"provider" mapstructure:"provider"`
	Protocol string            `json:"protocol" mapstructure:"protocol"`
	Metadata map[string]string `json:"metadata" mapstructure:"metadata"`
}

// CancelCmd 服务下线参数
type CancelCmd struct {
	EnvId       string `json:"envId" mapstructure:"envId"`
	AppId       string `json:"appId" mapstructure:"appId"`
	InstanceId  string `json:"instanceId" mapstructure:"instanceId"`
	LatestTime  int64  `json:"lastTime" mapstructure:"lastTime"`       //other node send
	Replication bool   `json:"replication" mapstructure:"replication"` //other node send
}

// FetchQry 服务发现参数
type FetchQry struct {
	EnvId  string `json:"envId" mapstructure:"envId"`
	AppId  string `json:"appId" mapstructure:"appId"`
	Status uint32 `json:"status" mapstructure:"status"`
}

// BatchFetchQry 批量服务发现参数
type BatchFetchQry struct {
	EnvId  string   `json:"envId" mapstructure:"envId"`
	AppId  []string `json:"appId" mapstructure:"appId"`
	Status uint32   `json:"status" mapstructure:"status"`
}
