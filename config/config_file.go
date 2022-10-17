package config

import (
	"github.com/spf13/viper"
	"runtime"
	"strings"
)

func FromFile(m *Manager) {
	//realPath, _ := filepath.Abs("./")
	// realPath := getCurrentDir()
	///configFilePath := realPath + "/" + env + "/"
	configFilePath := "./config/" + m.Options.Env + "/"
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(configFilePath)
	err := viper.ReadInConfig()
	if err != nil {
		realPath := getCurrentDir()
		configFilePath = realPath + "/" + m.Options.Env + "/"
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
		viper.AddConfigPath(configFilePath)
		err := viper.ReadInConfig()
		if err != nil {
			panic(err)
		}
	}
	_ = viper.Unmarshal(m.Bean)
}

func getCurrentDir() string {
	_, fileName, _, _ := runtime.Caller(1)
	aPath := strings.Split(fileName, "/")
	dir := strings.Join(aPath[0:len(aPath)-1], "/")
	return dir
}