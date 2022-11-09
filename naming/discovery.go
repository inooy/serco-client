package naming

import (
	"errors"
	"github.com/inooy/serco-client/pkg/log"
	"github.com/mitchellh/mapstructure"
)

func (m *ServiceManager) Fetch(req FetchQry) (*FetchData, error) {
	result, err := m.Client.RequestTcp("/api/naming/fetch", req, 3000)
	if err != nil {
		return nil, err
	}
	if result.Code != 200 {
		log.Errorf("config center poll fail: code=%d ,msg=%s", result.Code, result.Msg)
		return nil, errors.New(result.Msg)
	}
	log.Info("discovery service success")
	var data FetchData
	// 将 map 转换为指定的结构体
	if err = mapstructure.Decode(result.Data, &data); err != nil {
		return nil, err
	}
	return &data, nil
}
