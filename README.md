
# 🎉 serco-client 

- serco: 服务协调者
- serco-client: 客户端

# 🎊 TODO List 
1. [x] 提供shutdown接口
2. [x] 补充使用文档
3. [x] 配置变更改用event监听
4. [ ] 提供批量服务发现接口
5. [ ] 提供日志接口封装

# 🎯 RoadMap
1. [x] 心跳检测与断线重连
2. [x] 周期性轮询配置是否更新
3. [x] 提供配置监听
4. [ ] 支持集群
5. [ ] 故障转移
6. [ ] 分布式一致性

# 💯 使用 
## 引入依赖：
```shell
go get github.com/inooy/core-client
```

## 更新依赖
```shell
go get github.com/inooy/core-client@v0.4.2
```

## 编程使用
### 配置中心使用
```go
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
```

### 注册中心使用

provider:

```go

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
		EnvId:        "dev",
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

```

consumer:

```go
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
		log.Info(fmt.Sprintf("appid:%s,envId:%s,hostname:%s,addrs:%s\n",
			instance.AppId,
			instance.EnvId,
			instance.InstanceId,
			strings.Join(instance.Addrs, " ")))
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
```
