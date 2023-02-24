package main

import (
	"context"
	"fmt"
	"github.com/inooy/serco-client/core"
	"github.com/inooy/serco-client/internal/common"
	"github.com/inooy/serco-client/naming"
	"github.com/inooy/serco-client/pkg/log"
	"github.com/inooy/serco-client/serco"
	"strings"
)

type CustomConfig struct {
	Name string `json:"name"`
}

func main() {
	// 配置bean，配置信息会绑定到bean中，配置刷新时，bean属性会一起刷新
	conf := CustomConfig{}
	manager := serco.NewSerco(serco.Options{
		AppName:      "serco-consumer",
		EnvId:        "dev",
		RemoteAddr:   "127.0.0.1:9011",
		PollInterval: 120000,
		InstanceId:   "",
	})
	manager.SetupConfig(&conf)

	// AppId: "core-provider", EnvId: "dev", Hostname: "core.provider",
	manager.Client.On(core.NamespaceDiscovery, func(dto *core.EventDTO) {
		log.Infof("收到事件%+v", dto)
	})

	manager.SetupDiscovery(serco.RegistryOpts{})
	provider := naming.SubscribeProvider{
		Protocol: "http",
		Provider: "core-provider",
	}

	// 订阅服务
	err := manager.Subscribe([]*naming.SubscribeProvider{&provider})
	if err != nil {
		log.Error(err)
		return
	}

	rs, err := manager.GetInstance("core-provider")
	if err != nil {
		log.Error(err)
		return
	}
	for _, instance := range rs {
		log.Info(fmt.Sprintf("appId:%s,envId:%s,hostname:%s,addrs:%s\n",
			instance.AppId,
			instance.EnvId,
			instance.InstanceId,
			strings.Join(instance.AddressList, " ")))
	}
	fmt.Println("config name=" + conf.Name)

	common.GraceShutdown(func(ctx context.Context) {
		err = manager.Shutdown()
		if err != nil {
			fmt.Println(err)
		}
		ctx.Done()
	})
}
