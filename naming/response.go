package naming

type Instance struct {
	Env        string   `json:"env"`
	AppId      string   `json:"appId"`
	InstanceId string   `json:"instanceId"`
	Addrs      []string `json:"addrs"`
	Version    string   `json:"version"`
	Status     uint32   `json:"status"`

	RegTimestamp    int64 `json:"regTimestamp"`
	UpTimestamp     int64 `json:"upTimestamp"`
	RenewTimestamp  int64 `json:"renewTimestamp"`
	DirtyTimestamp  int64 `json:"dirtyTimestamp"`
	LatestTimestamp int64 `json:"latestTimestamp"`
}

type FetchData struct {
	Instances       []*Instance `json:"instances"`
	LatestTimestamp int64       `json:"latestTimestamp"`
}
