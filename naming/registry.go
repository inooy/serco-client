package naming

import (
	"errors"
	"github.com/inooy/serco-client/pkg/log"
	"net"
	"strconv"
	"strings"
	"time"
)

// GetIpAddr 获取本地IP地址 （问题：不能确保获取的ip的正确性）
func GetIpAddr() string {
	result := "127.0.0.1"
	addrArr, err := net.InterfaceAddrs()
	if err != nil {
		return result
	}

	for _, address := range addrArr {
		// 检查ip地址判断是否回环地址
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if cur := ipNet.IP.To4(); cur != nil {
				result = ipNet.IP.String()
				if strings.HasPrefix(result, "172") {
					return result
				}
			}
		}
	}
	return result
}

func buildAddr(protocol string, ip string, port int) string {
	return protocol + "://" + ip + ":" + strconv.Itoa(port)
}

func (m *ServiceManager) Registry() error {
	log.Info("start registry service")
	addr := ""
	if m.Options.LocalIp == "" {
		addr = buildAddr(m.Options.Protocol, GetIpAddr(), m.Options.LocalPort)
	} else {
		addr = buildAddr(m.Options.Protocol, m.Options.LocalIp, m.Options.LocalPort)
	}
	var req = RegisterCmd{
		AppId:      m.Options.AppName,
		Env:        m.Options.EnvType,
		InstanceId: m.Options.InstanceId,
		Addrs:      []string{addr},
		Status:     1,
	}
	m.Registered = true
	return m.RegistryRequest(req)
}

func (m *ServiceManager) Cancel() error {
	if m.Client == nil {
		return nil
	}
	var req = CancelCmd{
		AppId:           m.Options.AppName,
		Env:             m.Options.EnvType,
		InstanceId:      m.Options.InstanceId,
		LatestTimestamp: time.Now().UnixNano() / 1e6,
		Replication:     false,
	}
	return m.CancelRequest(req)
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
