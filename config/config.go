package config

import (
	"github.com/inooy/serco-client/config/remote"
	"github.com/inooy/serco-client/pkg/log"
	"strings"
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

type Status string

const (
	NotInit  Status = "NotInit"
	Starting Status = "Starting"
	Started  Status = "Started"
	Stopping Status = "Stopping"
	Stopped  Status = "Stopped"
)

type Manager struct {
	Options      Options
	Bean         interface{}
	emitter      PropEventEmitter
	MetadataList map[string]*remote.Metadata
	Status       Status
}

func NewManager(bean interface{}) *Manager {
	return &Manager{
		Status:       NotInit,
		MetadataList: map[string]*remote.Metadata{},
		emitter: PropEventEmitter{
			callbacks: make(map[string]func([]*PropChangeEvent)),
		},
		Bean: bean,
	}
}

func (m *Manager) OnFileChange(metadata *remote.Metadata) {
	UpdateConfigBean(metadata, m)
	events, err := calcChange(m.MetadataList[metadata.FileId], metadata)
	if err != nil {
		log.Error("calc config change error:", err.Error())
	} else {
		m.publishEvent(events)
	}
	m.MetadataList[metadata.FileId] = metadata
}

func (m *Manager) publishEvent(events []*PropChangeEvent) {
	for key := range m.emitter.callbacks {
		list := make([]*PropChangeEvent, 0)
		for i := range events {
			if strings.HasPrefix(events[i].PropName, key) {
				list = append(list, events[i])
			}
		}
		if len(list) > 0 {
			m.emitter.Emit(key, list)
		}
	}
}

func (m *Manager) On(id string, callback func([]*PropChangeEvent)) {
	m.emitter.On(id, callback)
}

// InitConfig init config, need invoke when app launch
func (m *Manager) InitConfig() {
	once.Do(func() {
		m.Status = Starting
		if m.Options.RemoteAddr == "" {
			FromFile(m)
		} else {
			FromServer(m)
		}
		m.Status = Started
	})
}

func (m *Manager) Shutdown() {
	if m.Status == Started {
		// todo close
		m.Status = Stopped
	}
}
