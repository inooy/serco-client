package client

import (
	"github.com/inooy/serco-client/pkg/log"
	"github.com/inooy/serco-client/pkg/socket/abilities"
	"github.com/inooy/serco-client/pkg/socket/connection"
	"github.com/inooy/serco-client/pkg/socket/model"
)

type SocketClient interface {
	Mount()
	Send(frame model.Frame) error
	SendHeartbeat() error
	Close(err error) error
	Connect() error
	ReConnect(err error) error
	IsConnect() bool
}

type Implement interface {
	GetHeartbeatFrame() model.Frame
}

type Template struct {
	Implement
	socketConnection connection.SocketConnection
	reconnectManager *abilities.ReconnectManager
	heartbeatManager *abilities.HeartbeatManager
}

func NewTemplate(impl Implement, socketConnection connection.SocketConnection) *Template {
	return &Template{
		Implement:        impl,
		socketConnection: socketConnection,
	}
}

func (t *Template) Mount() {
	t.reconnectManager = abilities.NewReconnectManager(t)
	t.heartbeatManager = abilities.NewHeartbeatManager(t)
	t.socketConnection.AddListener(connection.Listener{
		OnStatusChange: func(status connection.Status) {
			log.Info("socket status change:", status)
			// 扩展心跳机制
			if status == connection.CONNECTED {
				t.heartbeatManager.Pulse()
				t.reconnectManager.OnSocketConnectionSuccess()
			} else if status == connection.CLOSED || status == connection.ERROR {
				t.heartbeatManager.Dead()
			}
		},
		OnError: func(err error) {
			log.Error("socket error:", err.Error())
			t.reconnectManager.OnSocketDisconnection(err)
		},
	})
}

func (t *Template) ReceiveHeartbeat() {
	log.Info("收到心跳...")
	t.heartbeatManager.Feed()
}

func (t *Template) Send(frame model.Frame) error {
	return t.socketConnection.Send(frame)
}

func (t *Template) SendHeartbeat() error {
	return t.Send(t.GetHeartbeatFrame())
}

func (t *Template) Close(err error) error {
	t.heartbeatManager.Dead()
	t.reconnectManager.Shutdown()
	return t.socketConnection.Close(err)
}

func (t *Template) Connect() (err error) {
	err = t.socketConnection.Connect()
	t.reconnectManager.Start()
	return
}

func (t *Template) ReConnect(err error) error {
	return t.socketConnection.Close(err)
}

func (t *Template) IsConnect() bool {
	return t.socketConnection.GetConnectionState() == connection.CONNECTED
}
