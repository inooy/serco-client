package naming

import (
	"errors"
	"github.com/inooy/serco-client/pkg/log"
)

func (m *ServiceManager) Registry() error {
	var req = RegisterCmd{
		AppId:      m.AppId,
		Env:        m.Env,
		InstanceId: m.InstanceId,
		Addrs:      []string{"http://1.1.1.1"},
		Status:     1,
	}
	return m.RegistryRequest(req)
}

func (m *ServiceManager) RegistryRequest(req RegisterCmd) error {
	result, err := m.Client.RequestTcp("/api/naming/registry", req, 3000)
	if err != nil {
		return err
	}
	if result.Code != 200 {
		log.Errorf("config center poll fail: code=%d ,msg=%s", result.Code, result.Msg)
		return errors.New(result.Msg)
	}
	log.Info("注册成功")
	return nil
}

func (m *ServiceManager) SubscribeRequest(req SubscribeCmd) error {
	result, err := m.Client.RequestTcp("/api/naming/subscribe", req, 3000)
	if err != nil {
		return err
	}
	if result.Code != 200 {
		log.Errorf("config center poll fail: code=%d ,msg=%s", result.Code, result.Msg)
		return errors.New(result.Msg)
	}
	log.Info("subscribe service success")
	return nil
}

func (m *ServiceManager) CancelRequest(req CancelCmd) error {
	result, err := m.Client.RequestTcp("/api/naming/cancel", req, 3000)
	if err != nil {
		return err
	}
	if result.Code != 200 {
		log.Errorf("config center poll fail: code=%d ,msg=%s", result.Code, result.Msg)
		return errors.New(result.Msg)
	}
	log.Info("注销成功")
	return nil
}
