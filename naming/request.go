package naming

// RegisterCmd 服务注册参数
type RegisterCmd struct {
	Env        string `json:"env" mapstructure:"env"`
	AppId      string `json:"appId" mapstructure:"appId"`
	InstanceId string `json:"instanceId" mapstructure:"instanceId"`
	// 服务实例的地址，可以是 http 或 rpc 地址，多个地址可以维护数组
	Addrs           []string `json:"addrs" mapstructure:"addrs"`
	Status          uint32   `json:"status" mapstructure:"status"`
	Version         string   `json:"version" mapstructure:"version"`
	LatestTimestamp int64    `json:"latestTimestamp" mapstructure:"latestTimestamp"`
	DirtyTimestamp  int64    `json:"dirtyTimestamp" mapstructure:"dirtyTimestamp"` //other node send
	Replication     bool     `json:"replication" mapstructure:"replication"`       //other node send
}

// SubscribeCmd 服务订阅参数
type SubscribeCmd struct {
	InstanceId string              `json:"instanceId" mapstructure:"instanceId"`
	Subscribes []SubscribeProvider `json:"subscribes" mapstructure:"subscribes"`
}

type SubscribeProvider struct {
	Provider string            `json:"provider" mapstructure:"provider"`
	Protocol string            `json:"protocol" mapstructure:"protocol"`
	Meta     map[string]string `json:"meta" mapstructure:"meta"`
}

// CancelCmd 服务下线参数
type CancelCmd struct {
	Env             string `json:"env" mapstructure:"env"`
	AppId           string `json:"appId" mapstructure:"appId"`
	InstanceId      string `json:"instanceId" mapstructure:"instanceId"`
	LatestTimestamp int64  `json:"lastTimestamp" mapstructure:"lastTimestamp"` //other node send
	Replication     bool   `json:"replication" mapstructure:"replication"`     //other node send
}

// FetchQry 服务发现参数
type FetchQry struct {
	Env    string `json:"env" mapstructure:"env"`
	AppId  string `json:"appId" mapstructure:"appId"`
	Status uint32 `json:"status" mapstructure:"status"`
}

// BatchFetchQry 批量服务发现参数
type BatchFetchQry struct {
	Env    string   `json:"env" mapstructure:"env"`
	AppId  []string `json:"appId" mapstructure:"appId"`
	Status uint32   `json:"status" mapstructure:"status"`
}
