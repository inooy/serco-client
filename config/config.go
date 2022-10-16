package config

import (
	"github.com/inooy/serco-client/config/load"
	"github.com/inooy/serco-client/config/remote"
	"github.com/inooy/serco-client/pkg/log"
	"github.com/inooy/serco-client/pkg/socket/connection"
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

func (m *Manager) InitConfig() {
	once.Do(func() {
		if m.Options.RemoteAddr == "" {
			load.FromFile(m)
		} else {
			load.FromServer(m)
			conn := remote.NewConfigSocketClient(connection.TcpSocketConnectOpts{
				Host: m.Options.RemoteAddr,
			})
			conn.Mount()
			err := conn.Connect()
			if err != nil {
				panic(err)
			}
			result, err := conn.RequestTcp("/app/config", "", 12)
			if err != nil {
				panic(err)
			}
			log.Info(result)
		}
	})
}
