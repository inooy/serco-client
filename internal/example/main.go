package main

import (
	"fmt"
	"github.com/inooy/serco-client/config"
	"github.com/inooy/serco-client/pkg/log"
	"time"
)

type CustomConfig struct {
	Name string `json:"name"`
}

func main() {
	// 配置bean，配置信息会绑定到bean中，配置刷新时，bean属性会一起刷新
	conf := CustomConfig{}
	// 构造配置管理器
	configManager := config.NewManager(config.Options{
		AppName:      "core-demo",
		Env:          "dev",
		RemoteAddr:   "127.0.0.1:9011",
		PollInterval: 300000,
	}, &conf)
	configManager.InitConfig()
	// 监听配置修改
	configManager.On("app.name", func(events []*config.PropChangeEvent) {
		log.Infof("监听到配置属性更新, key=app.name, list len=%d", len(events))
		for i := range events {
			log.Infof("配置change event: %+v", events[i])
		}
	})
	fmt.Println("config name=" + conf.Name)
	time.Sleep(5 * time.Minute)
	fmt.Println("refreshed config name=" + conf.Name)
	configManager.Shutdown()
}
