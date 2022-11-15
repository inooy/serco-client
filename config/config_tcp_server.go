package config

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/inooy/serco-client/core"
	"github.com/inooy/serco-client/pkg/log"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"time"
)

type MetadataChangeEvent struct {
	AppName  string    `json:"appName" mapstructure:"appName"`
	EnvType  string    `json:"envType" mapstructure:"envType"`
	Metadata *Metadata `json:"metadata"`
}

type CheckRequest struct {
	AppName string         `json:"appName" mapstructure:"appName"`
	EnvType string         `json:"envType" mapstructure:"envType"`
	Old     map[string]int `json:"old" mapstructure:"old"`
}

func (m *Manager) FromServer() {
	log.Info("从配置中心获取配置信息")
	m.Client.On(core.NamespaceConfig, func(dto *core.EventDTO) {
		var data MetadataChangeEvent
		//将 map 转换为指定的结构体
		if err := mapstructure.Decode(dto.Data, &data); err != nil {
			fmt.Println(err)
		}
		m.OnFileChange(data.Metadata)
	})
	m.Client.OnReconnected(func(isReconnect bool) error {
		if !isReconnect {
			return nil
		}
		log.Info("reconnect success, start re login...")
		result, err := m.Client.Login(m.Options.AppName, m.Options.Env, 6000)
		if err != nil {
			return err
		}
		if result.Code != 200 {
			return errors.New(result.Msg)
		}
		log.Info("re login success!")
		var list []Metadata
		//将 map 转换为指定的结构体
		if err = mapstructure.Decode(result.Data, &list); err != nil {
			return err
		}
		for i := range list {
			m.OnFileChange(&list[i])
		}
		return nil
	})
	result, err := m.Client.Launch(m.Options.AppName, m.Options.Env, 6000)
	if err != nil {
		panic(err)
	}
	log.Info(result)
	if result.Code != 200 {
		panic(result.Msg)
	}
	var list []Metadata
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
	for i := range list {
		log.Info("加载到配置中心配置:", list[i].FileId)
		log.Info("配置内容：", list[i].Content)
		m.MetadataList[list[i].FileId] = &list[i]
		err := con.ReadConfig(bytes.NewBufferString(list[i].Content))
		if err != nil {
			panic(err)
		}
		err = con.Unmarshal(m.Bean)
		if err != nil {
			panic(err)
		}
	}
	go startPoll(m)
}

func startPoll(m *Manager) {
	defer func() {
		log.Info("poll config exit")
	}()
	log.Info("start config center poll")
	for {
		if m.Status == Stopping || m.Status == Stopped {
			log.Info("stop config poll")
			return
		}
		var interval = 120000
		if m.Options.PollInterval > 0 {
			interval = m.Options.PollInterval
		}
		time.Sleep(time.Duration(interval) * time.Millisecond)
		old := make(map[string]int)
		for s := range m.MetadataList {
			old[m.MetadataList[s].FileId] = m.MetadataList[s].Version
		}
		req := CheckRequest{
			AppName: m.Options.AppName,
			EnvType: m.Options.Env,
			Old:     old,
		}
		result, err := m.Client.RequestTcp("/api/config/check", req, 3000)
		if err != nil {
			log.Error("config center poll error!", err.Error())
			continue
		}
		if result.Code != 200 {
			log.Errorf("config center poll fail: code=%d ,msg=%s", result.Code, result.Msg)
			continue
		}
		var list []Metadata
		//将 map 转换为指定的结构体
		if err = mapstructure.Decode(result.Data, &list); err != nil {
			log.Error("config center polled result error env="+m.Options.Env+"appName="+m.Options.AppName+",server="+m.Options.RemoteAddr, err.Error())
			continue
		}
		log.Info("poll config success! count: ", len(list))
		for i := range list {
			m.OnFileChange(&list[i])
		}
	}
}
