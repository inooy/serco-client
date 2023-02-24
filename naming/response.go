package naming

type Instance struct {
	EnvId        string   `json:"envId"`
	AppId        string   `json:"appId"`
	InstanceId   string   `json:"instanceId"`
	AddressList  []string `json:"addressList"`
	Version      string   `json:"version"`
	Status       uint32   `json:"status"`
	RegisterTime int64    `json:"registerTime"`
	UpTime       int64    `json:"upTime"`
	RenewTime    int64    `json:"renewTime"`
	DirtyTime    int64    `json:"dirtyTime"`
	LatestTime   int64    `json:"latestTime"`
}

type FetchData struct {
	Instances  []*Instance `json:"instances"`
	LatestTime int64       `json:"latestTime"`
}
