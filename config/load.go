package config

import (
	"bytes"
	"github.com/inooy/serco-client/config/remote"
	"github.com/inooy/serco-client/pkg/log"
	"github.com/spf13/viper"
)

func UpdateConfigBean(metadata *remote.Metadata, m *Manager) {
	log.Info("准别更新配置")
	con := viper.New()
	con.SetConfigType("yaml")
	con.SetConfigName("config")
	log.Info("加载到配置中心配置:", metadata.FileId)
	log.Info("配置内容：", metadata.Content)
	err := con.ReadConfig(bytes.NewBufferString(metadata.Content))
	if err != nil {
		log.Error("配置更新失败：", err.Error())
		return
	}
	err = con.Unmarshal(m.Bean)
	if err != nil {
		log.Error("配置绑定失败：", err.Error())
	}
}
