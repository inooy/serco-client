package main

import (
	"context"
	"fmt"
	"github.com/inooy/serco-client/internal/common"
	"github.com/inooy/serco-client/serco"
)

func main() {
	// 构造配置管理器
	manager := serco.NewSerco(serco.Options{
		AppName:      "core-provider",
		Env:          "dev",
		RemoteAddr:   "127.0.0.1:9011",
		PollInterval: 300000,
		InstanceId:   "",
	})
	manager.SetupDiscovery(serco.RegistryOpts{
		LocalIp:   "",
		LocalPort: 9090,
		Protocol:  "http",
	})

	err := manager.Registry()
	if err != nil {
		panic(err)
	}

	fmt.Println("config name=" + manager.Options.AppName)
	common.GraceShutdown(func(ctx context.Context) {
		err = manager.Shutdown()
		if err != nil {
			fmt.Println(err)
		}
	})
}
