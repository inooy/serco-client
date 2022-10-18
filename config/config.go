package config

import (
	"github.com/inooy/serco-client/config/remote"
	"sync"
)

var once sync.Once

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

type Manager struct {
	Options      Options
	Bean         interface{}
	MetadataList map[string]*remote.Metadata
}

func (m *Manager) OnFileChange(metadata *remote.Metadata) {
	UpdateConfigBean(metadata, m)
	m.MetadataList[metadata.FileId] = metadata
}

// InitConfig init config, need invoke when app launch
func (m *Manager) InitConfig() {
	m.MetadataList = map[string]*remote.Metadata{}
	once.Do(func() {
		if m.Options.RemoteAddr == "" {
			FromFile(m)
		} else {
			FromServer(m)
		}
	})
}
