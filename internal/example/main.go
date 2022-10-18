package main

import (
	"fmt"
	"github.com/inooy/serco-client/config"
	"time"
)

type CustomConfig struct {
	Name string `json:"name"`
}

func main() {
	conf := CustomConfig{}
	configManager := config.Manager{
		Options: config.Options{
			AppName:      "serco-demo",
			Env:          "dev",
			RemoteAddr:   "127.0.0.1:9011",
			PollInterval: 300000,
		},
		Bean: &conf,
	}
	configManager.InitConfig()
	fmt.Println("config name=" + conf.Name)
	time.Sleep(5 * time.Minute)
	fmt.Println("refreshed config name=" + conf.Name)

}
