package main

import (
	"fmt"
	"github.com/inooy/serco-client/config"
	"github.com/inooy/serco-client/pkg/log"
	"github.com/inooy/serco-client/serco"
	"time"
)

type CustomConfig struct {
	Name string `json:"name"`
}

func main() {
	// 配置bean，配置信息会绑定到bean中，配置刷新时，bean属性会一起刷新
	conf := CustomConfig{}

	Serco := serco.NewSerco(serco.Options{
		AppName:      "core-demo",
		EnvId:        "dev",
		RemoteAddr:   "127.0.0.1:9011",
		PollInterval: 120000,
		InstanceId:   "",
	})
	Serco.SetupConfig(&conf)
	// 监听配置修改
	Serco.ConfigManager.On("app.name", func(events []*config.PropChangeEvent) {
		log.Infof("监听到配置属性更新, key=app.name, list len=%d", len(events))
		for i := range events {
			log.Infof("配置change event: %+v", events[i])
		}
	})
	fmt.Println("config name=" + conf.Name)
	time.Sleep(5 * time.Minute)
	fmt.Println("refreshed config name=" + conf.Name)
	// 优雅关闭
	err := Serco.Shutdown()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
