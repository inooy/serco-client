package client

import (
	"context"
	"fmt"
	"github.com/inooy/serco-client/pkg/log"
	"github.com/inooy/serco-client/pkg/socket/abilities"
	"github.com/inooy/serco-client/pkg/socket/connection"
	"github.com/inooy/serco-client/pkg/socket/model"
	"github.com/inooy/serco-client/pkg/tools"
	"sync"
	"time"
)

type SocketClient interface {
	Mount()
	Send(frame model.Frame) error
	SendHeartbeat() error
	Close(err error) error
	Connect() error
	ReConnect(err error) error
	IsConnect() bool
	RequestTcp(path string, content interface{}, timeout int) (interface{}, error)
}

type Implement interface {
	GetHeartbeatFrame() model.Frame
}

type Template struct {
	Implement
	Emitter          *RpcEventEmitter
	socketConnection connection.SocketConnection
	reconnectManager *abilities.ReconnectManager
	heartbeatManager *abilities.HeartbeatManager
}

func NewTemplate(impl Implement, socketConnection connection.SocketConnection) *Template {
	return &Template{
		Implement:        impl,
		socketConnection: socketConnection,
		Emitter:          &RpcEventEmitter{},
	}
}

type RpcEventEmitter struct {
	cLock     sync.RWMutex // protect the map
	callbacks map[string]func(*ResponseDTO)
}

func (emitter *RpcEventEmitter) Once(id string, callback func(*ResponseDTO)) {
	emitter.cLock.Lock()
	emitter.callbacks[id] = func(dto *ResponseDTO) {
		emitter.Off(id)
		callback(dto)
	}
	emitter.cLock.Unlock()
}

func (emitter *RpcEventEmitter) Off(id string) {
	emitter.cLock.RLock()
	if _, ok := emitter.callbacks[id]; ok {
		delete(emitter.callbacks, id)
	}
	emitter.cLock.RUnlock()
}

func (emitter *RpcEventEmitter) Emit(id string, dto *ResponseDTO) {
	if callback, ok := emitter.callbacks[id]; ok {
		callback(dto)
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

func (t *Template) RequestTcp(path string, content interface{}, timeout int) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	var result interface{}
	var err error
	// 每个请求生成唯一的请求id，超时移除对应回调监听
	seq := tools.GetSnowflakeId()

	go func(ctx context.Context) {
		// 发送HTTP请求
		t.Emitter.Once(seq, func(dto *ResponseDTO) {
			result = dto.Body
			cancel()
		})

		requestDTO := RequestDTO{
			Body: content,
		}
		requestDTO.Header = RequestHeader{
			Path: path,
		}
		requestDTO.Header.SubSeq = seq
		packet := model.PacketFrame{Cmd: model.CommandRequest, Body: requestDTO}
		err = t.Send(packet)
		if err != nil {
			log.Errorf("发送失败: %s", err.Error())
			cancel()
		}
	}(ctx)

	select {
	case <-ctx.Done():
		fmt.Println("call successfully!!!")
		return result, err
	}
}
