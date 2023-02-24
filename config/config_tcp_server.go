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
	EnvId    string    `json:"envId" mapstructure:"envId"`
	Metadata *Metadata `json:"metadata"`
}

type CheckRequest struct {
	AppName string         `json:"appName" mapstructure:"appName"`
	EnvId   string         `json:"envId" mapstructure:"envId"`
	Old     map[string]int `json:"old" mapstructure:"old"`
}

type SubscribeRequest struct {
	AppName    string `json:"appName" mapstructure:"appName"`
	EnvId      string `json:"envId" mapstructure:"envId"`
	InstanceId string `json:"instanceId" mapstructure:"instanceId"`
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
		list, err := m.subscribe()
		if err != nil {
			return err
		}
		for i := range list {
			m.OnFileChange(&list[i])
		}
		return nil
	})
	list, err := m.subscribe()

	if err != nil {
		panic(err)
	}

	if len(list) == 0 {
		panic("配置中心无配置env=" + m.Options.EnvId + "appName=" + m.Options.AppName + ",server=" + m.Options.RemoteAddr)
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

func (m *Manager) subscribe() ([]Metadata, error) {
	req := SubscribeRequest{
		AppName: m.Options.AppName,
		EnvId:   m.Options.EnvId,
	}
	result, err := m.Client.RequestTcp("/api/config/subscribe", req, 3000)
	if err != nil {
		return nil, err
	}
	if result.Code != 200 {
		return nil, errors.New(fmt.Sprintf("config center poll fail: code=%d ,msg=%s", result.Code, result.Msg))
	}
	var list []Metadata
	//将 map 转换为指定的结构体
	if err = mapstructure.Decode(result.Data, &list); err != nil {
		return nil, errors.New("配置中心无配置envId=" + m.Options.EnvId + "appName=" + m.Options.AppName + ",server=" + m.Options.RemoteAddr)
	}
	return list, nil
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
			EnvId:   m.Options.EnvId,
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
			log.Error("config center polled result error envId="+m.Options.EnvId+"appName="+m.Options.AppName+",server="+m.Options.RemoteAddr, err.Error())
			continue
		}
		log.Info("poll config success! count: ", len(list))
		for i := range list {
			m.OnFileChange(&list[i])
		}
	}
}
