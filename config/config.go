package config

import (
	"github.com/inooy/serco-client/config/remote"
	"sync"
)

var once sync.Once

type Options struct {
	Env        string
	AppName    string
	RemoteAddr string
}

type Manager struct {
	Options Options
	Bean    interface{}
}

func (m *Manager) OnFileChange(metadata *remote.Metadata) {
	UpdateConfigBean(metadata, m)
}

func (m *Manager) InitConfig() {
	once.Do(func() {
		if m.Options.RemoteAddr == "" {
			FromFile(m)
		} else {
			FromServer(m)
		}
	})
}
