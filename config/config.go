package config

import (
	"bytes"
	"github.com/inooy/serco-client/core"
	"github.com/inooy/serco-client/pkg/log"
	"github.com/spf13/viper"
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
	Options      *Options
	Bean         interface{}
	emitter      PropEventEmitter
	MetadataList map[string]*Metadata
	Status       Status
	Client       *core.SocketClientImpl
}

func NewManager(options *Options, bean interface{}, client *core.SocketClientImpl) *Manager {
	return &Manager{
		Status:       NotInit,
		MetadataList: map[string]*Metadata{},
		emitter: PropEventEmitter{
			callbacks: make(map[string]func([]*PropChangeEvent)),
		},
		Options: options,
		Client:  client,
		Bean:    bean,
	}
}

func (m *Manager) OnFileChange(metadata *Metadata) {
	m.UpdateConfigBean(metadata)
	events, err := calcChange(m.MetadataList[metadata.FileId], metadata)
	if err != nil {
		log.Error("calc config change error:", err.Error())
	} else {
		m.publishEvent(events)
	}
	m.MetadataList[metadata.FileId] = metadata
}

func (m *Manager) UpdateConfigBean(metadata *Metadata) {
	log.Info("准备更新配置")
	con := viper.New()
	con.SetConfigType("yaml")
	con.SetConfigName("config")
	log.Info("加载到配置中心配置:", metadata.FileId)
	log.Info("配置内容：", metadata.Content)
	err := con.ReadConfig(bytes.NewBufferString(metadata.Content))
	if err != nil {
		log.Error("配置更新失败：", err.Error())
		return
	}
	err = con.Unmarshal(m.Bean)
	if err != nil {
		log.Error("配置绑定失败：", err.Error())
	}
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

// On 监听配置变更，id为配置项， such as: app.name
func (m *Manager) On(id string, callback func([]*PropChangeEvent)) {
	m.emitter.On(id, callback)
}

// InitConfig init config, need invoke when app launch
func (m *Manager) InitConfig() {
	once.Do(func() {
		m.Status = Starting
		if m.Options.RemoteAddr == "" {
			m.FromFile()
		} else {
			m.FromServer()
		}
		m.Status = Started
	})
}

// Shutdown 优雅下线配置中心连接
func (m *Manager) Shutdown() {
	if m.Status == Started {
		log.Info("start shutdown config")
		m.Status = Stopping
		if m.Client != nil {
			_ = m.Client.Close(nil)
		}
		m.Status = Stopped
		log.Info("config shutdown finished!")
	}
}
