package main

import (
	"fmt"
	"github.com/inooy/serco-client/config"
	"github.com/inooy/serco-client/config/remote"
	"github.com/inooy/serco-client/naming"
	"github.com/inooy/serco-client/pkg/log"
	"strings"
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
		AppName:      "serco-demo",
		Env:          "dev",
		RemoteAddr:   "127.0.0.1:9011",
		PollInterval: 300000,
	}, &conf)
	configManager.InitConfig()

	var req1 = naming.RegisterCmd{AppId: "serco-provider", Env: "dev", InstanceId: "serco.provider", Addrs: []string{"http://1.1.1.1/testapp"}, Status: 1}

	manager := naming.ServiceManager{Client: configManager.Client}

	// AppId: "serco-provider", Env: "dev", Hostname: "serco.provider",
	manager.Client.EventEmitter.On("serco-provider", func(dto *remote.EventDTO) {
		log.Infof("收到事件%+v", dto)
	})

	provider := naming.SubscribeProvider{
		Protocol: "http",
		Provider: "serco-provider",
	}
	subscribe := naming.SubscribeCmd{
		InstanceId: "serco-consumer1",
		Subscribes: []naming.SubscribeProvider{provider},
	}

	// 订阅服务
	err := manager.Subscribe(subscribe)
	if err != nil {
		log.Error(err)
		return
	}

	req := naming.FetchQry{
		Env:    req1.Env,
		AppId:  req1.AppId,
		Status: req1.Status,
	}
	rs, err := manager.Fetch(req)
	if err != nil {
		log.Error(err)
		return
	}
	for _, instance := range rs.Instances {
		log.Info(fmt.Sprintf("appid:%s,env:%s,hostname:%s,addrs:%s\n",
			instance.AppId,
			instance.Env,
			instance.InstanceId,
			strings.Join(instance.Addrs, " ")))
	}
	fmt.Println("config name=" + conf.Name)
	time.Sleep(5 * time.Minute)
	fmt.Println("refreshed config name=" + conf.Name)
	configManager.Shutdown()
}
