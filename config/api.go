package config

type Metadata struct {
	Id         int    `json:"id"`
	AppName    string `json:"appName"`
	EnvId      string `json:"envId"`
	FileId     string `json:"fileId"`
	Format     string `json:"format"`
	Content    string `json:"content"`
	CreateTime int64  `json:"createTime"`
	UpdateTime int64  `json:"updateTime"`
	Version    int    `json:"version"`
}

type FileListener interface {
	OnFileChange(metadata *Metadata)
}
