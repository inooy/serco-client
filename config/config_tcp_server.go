package config

import (
	"bytes"
	"github.com/inooy/serco-client/config/remote"
	"github.com/inooy/serco-client/pkg/log"
	"github.com/inooy/serco-client/pkg/socket/connection"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

func FromServer(m *Manager) {
	log.Info("从配置中心获取配置信息")

	conn := remote.NewConfigSocketClient(connection.TcpSocketConnectOpts{
		Host: m.Options.RemoteAddr,
	}, m)
	err := conn.Launch()
	if err != nil {
		panic(err)
	}
	result, err := conn.Login(m.Options.AppName, m.Options.Env, 6000)
	if err != nil {
		panic(err)
	}
	log.Info(result)
	if result.Code != 200 {
		panic(result.Msg)
	}
	var list []remote.Metadata
	//将 map 转换为指定的结构体
	if err = mapstructure.Decode(result.Data, &list); err != nil {
		panic("配置中心无配置env=" + m.Options.Env + "appName=" + m.Options.AppName + ",server=" + m.Options.RemoteAddr)
	}

	if len(list) == 0 {
		panic("配置中心无配置env=" + m.Options.Env + "appName=" + m.Options.AppName + ",server=" + m.Options.RemoteAddr)
	}
	con := viper.New()
	con.SetConfigType("yaml")
	con.SetConfigName("config")
	for _, metadata := range list {
		log.Info("加载到配置中心配置:", metadata.FileId)
		log.Info("配置内容：", metadata.Content)
		err := con.ReadConfig(bytes.NewBufferString(metadata.Content))
		if err != nil {
			panic(err)
		}
		err = con.Unmarshal(m.Bean)
		if err != nil {
			panic(err)
		}
	}
}
