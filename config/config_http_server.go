package config

import (
	"bytes"
	"errors"
	"github.com/inooy/serco-client/config/remote"
	"github.com/inooy/serco-client/pkg/log"
	"github.com/spf13/viper"
	"github.com/swxctx/ghttp"
)

func FromHttpServer(m *Manager) {
	log.Info("从配置中心获取配置信息")
	list, err := requestServer(m.Options.RemoteAddr, m.Options.Env, m.Options.AppName)
	if err != nil {
		panic(err)
	}
	if len(list) == 0 {
		panic("配置中心无配置env=" + m.Options.Env + "appName=" + m.Options.AppName + ",server=" + m.Options.RemoteAddr)
	}
	con := viper.New()
	con.SetConfigType("yaml")
	con.SetConfigName("config")
	for _, metadata := range list {
		log.Info("加载到配置中心配置:", metadata.FileId)
		log.Info("配置内容：", metadata.Content)
		err := con.ReadConfig(bytes.NewBufferString(metadata.Content))
		if err != nil {
			panic(err)
		}
		err = con.Unmarshal(m.Bean)
		if err != nil {
			panic(err)
		}
	}
}

type FetchConfigParams struct {
	AppName string `json:"appName"`
	EnvType string `json:"envType"`
}

type ResultRep struct {
	Code    int32             `json:"code"`
	Message string            `json:"message"`
	Data    []remote.Metadata `json:"data"`
}

func requestServer(confSrv string, env string, appName string) ([]remote.Metadata, error) {
	req := ghttp.Request{
		Url:         confSrv + "/config/metadata/list",
		Method:      "GET",
		ContentType: "application/json",
		Query: FetchConfigParams{
			AppName: appName,
			EnvType: env,
		},
	}
	log.Info("请求地址url=" + req.Url)
	resp, err := req.Do()
	if err != nil {
		return nil, err
	}
	defer func(Body *ghttp.Body) {
		err := Body.Close()
		if err != nil {
			log.Error("close fail" + err.Error())
		}
	}(resp.Body)

	var result ResultRep
	err = resp.Body.FromToJson(&result)
	if err != nil {
		return nil, err
	}
	if result.Code != 200 {
		return nil, errors.New(result.Message)
	}
	return result.Data, nil
}