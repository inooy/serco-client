
# 🎉 serco-client 

- serco: 服务协调者
- serco-client: 客户端

# 🎊 TODO List 
1. [ ] 提供shutdown接口
2. [ ] 提供日志接口封装
3. [ ] 补充使用文档

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
go get github.com/inooy/serco-client
```

## 更新依赖
```shell
go get github.com/inooy/serco-client@v0.1.1
```

## 编程使用

```go

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
		AppName:      "serco-demo",
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
}

```


