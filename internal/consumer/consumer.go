package main

import (
	"context"
	"fmt"
	"github.com/inooy/serco-client/config"
	"github.com/inooy/serco-client/config/remote"
	"github.com/inooy/serco-client/naming"
	"github.com/inooy/serco-client/pkg/log"
	"os"
	"os/signal"
	"strings"
	"syscall"
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
		PollInterval: 120000,
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

	// 等待中断信号来优雅地关闭服务器，为关闭服务器操作设置一个5秒的超时
	quit := make(chan os.Signal, 1) // 创建一个接收信号的通道
	// kill 默认会发送 syscall.SIGTERM 信号
	// kill -2 发送 syscall.SIGINT 信号，我们常用的Ctrl+C就是触发系统SIGINT信号
	// kill -9 发送 syscall.SIGKILL 信号，但是不能被捕获，所以不需要添加它
	// signal.Notify把收到的 syscall.SIGINT或syscall.SIGTERM 信号转发给quit
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT) // 此处不会阻塞
	<-quit                                                                                // 阻塞在此，当接收到上述两种信号时才会往下执行
	log.Info("Shutdown Server ...")
	// 创建一个5秒超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cancelReq := naming.CancelCmd{
		AppId:           req1.AppId,
		Env:             req1.Env,
		InstanceId:      req1.InstanceId,
		LatestTimestamp: time.Now().UnixNano(),
	}
	manager.Cancel(cancelReq)

	configManager.Shutdown()

	select {
	case <-ctx.Done():
		log.Warn("timeout of 10 seconds")
	}
	log.Info("server exiting")

	//time.Sleep(5 * time.Minute)
	fmt.Println("refreshed config name=" + conf.Name)
	configManager.Shutdown()
}
