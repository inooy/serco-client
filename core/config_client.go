package core

import (
	"errors"
	"fmt"
	"github.com/inooy/serco-client/pkg/log"
	"github.com/inooy/serco-client/pkg/socket/client"
	"github.com/inooy/serco-client/pkg/socket/connection"
	"github.com/inooy/serco-client/pkg/socket/model"
	"github.com/inooy/serco-client/pkg/tools"
	"github.com/mitchellh/mapstructure"
	"sync"
	"time"
)

type LoginRequest struct {
	SeqId   string `json:"seqId"` // sequence number chosen by client
	Token   string `json:"token"`
	AppName string `json:"appName"`
	EnvType string `json:"envType"`
}

type EventDTO struct {
	Id        string      `json:"id"`        // 事件id
	Topic     string      `json:"topic"`     // 主题
	Namespace string      `json:"namespace"` // 命名空间
	Data      interface{} `json:"data"`      // 事件数据
}

type SocketClientImpl struct {
	*client.Template
	loginCallback       func(*model.Response)
	reconnectedCallback []func(isReconnect bool) error
	heartBeatFrame      *model.HeartbeatFrame
	starting            bool
	emitter             EventEmitter
}

type EventEmitter struct {
	cLock     sync.RWMutex // protect the map
	callbacks map[string]func(*EventDTO)
}

func (s *SocketClientImpl) OnReconnected(callback func(isReconnect bool) error) {
	s.reconnectedCallback = append(s.reconnectedCallback, callback)
}

func (s *SocketClientImpl) On(namespace string, callback func(*EventDTO)) {
	s.emitter.cLock.Lock()
	s.emitter.callbacks[namespace] = callback
	s.emitter.cLock.Unlock()
}

func (s *SocketClientImpl) Off(namespace string) {
	s.emitter.cLock.Lock()
	if _, ok := s.emitter.callbacks[namespace]; ok {
		delete(s.emitter.callbacks, namespace)
	}
	s.emitter.cLock.Unlock()
}

func (s *SocketClientImpl) Emit(namespace string, dto *EventDTO) {
	if callback, ok := s.emitter.callbacks[namespace]; ok {
		callback(dto)
	}
}

func NewConfigSocketClient(options connection.TcpSocketConnectOpts) *SocketClientImpl {
	impl := &SocketClientImpl{
		emitter: EventEmitter{
			callbacks: make(map[string]func(*EventDTO)),
		},
		reconnectedCallback: make([]func(isReconnected bool) error, 0),
	}
	conn := connection.NewTcpConnection(options, &Codec{})
	template := client.NewTemplate(impl, conn)
	impl.Template = template
	conn.AddListener(connection.Listener{
		OnReceive: func(frame model.Frame) {
			impl.handleFrame(frame)
		},
		OnStatusChange: func(status connection.Status) {
			// 非启动时重连，那么重新登录
			if status == connection.CONNECTED {
				for i := range impl.reconnectedCallback {
					err := impl.reconnectedCallback[i](!impl.starting)
					if err != nil {
						_ = impl.Close(err)
						return
					}
				}
			}
		},
	})
	return impl
}

func (s *SocketClientImpl) handleFrame(frame model.Frame) {
	switch frame.(type) {
	case *model.HeartbeatFrame:
		s.ReceiveHeartbeat()
	case model.HeartbeatFrame:
		s.ReceiveHeartbeat()
		return
	}
	packet := frame.(model.PacketFrame)
	if packet.Cmd == model.CommandReply {
		var response model.ResponseDTO
		//将 map 转换为指定的结构体
		if err := mapstructure.Decode(packet.Body, &response); err != nil {
			fmt.Println(err)
		}
		// 发布事件
		s.Emitter.Emit(response.Header.SubSeq, &response)
	} else if packet.Cmd == model.CommandLoginResponse {
		// 登录响应
		var response model.Response
		//将 map 转换为指定的结构体
		if err := mapstructure.Decode(packet.Body, &response); err != nil {
			fmt.Println(err)
		}
		if s.loginCallback != nil {
			s.loginCallback(&response)
		}
	} else if packet.Cmd == model.CommandEvent {
		log.Infof("收到事件")
		// 事件机制，根据topic分发事件
		// topic subTopic
		var event EventDTO
		//将 map 转换为指定的结构体
		if err := mapstructure.Decode(packet.Body, &event); err != nil {
			fmt.Println(err)
		}
		s.Emit(event.Namespace, &event)
	}
}

func (s *SocketClientImpl) GetHeartbeatFrame() model.Frame {
	return s.heartBeatFrame
}

func (s *SocketClientImpl) SendData(cmd int, data interface{}) error {
	return s.Send(model.PacketFrame{
		Cmd:  model.Command(cmd),
		Body: data,
	})
}

var TimeoutErr = errors.New("timeout error")

func (s *SocketClientImpl) Login(appName string, envType string, timeout int) (*model.Response, error) {
	ch := make(chan *model.Response, 1)

	var err error
	// 每个请求生成唯一的请求id，超时移除对应回调监听
	seq := tools.GetSnowflakeId()

	go func() {
		s.loginCallback = func(response *model.Response) {
			ch <- response
		}
		requestDTO := LoginRequest{
			SeqId:   seq,
			AppName: appName,
			EnvType: envType,
		}
		packet := model.PacketFrame{Cmd: model.CommandLogin, Body: requestDTO}
		err = s.Send(packet)
		if err != nil {
			log.Errorf("发送失败: %s", err.Error())
			ch <- nil
		}
	}()
	var result *model.Response

	select {
	case <-time.After(time.Duration(timeout) * time.Millisecond):
		return nil, TimeoutErr
	case result = <-ch:
		return result, err
	}
}

func (s *SocketClientImpl) Launch(appName string, envType string, timeout int) (*model.Response, error) {
	s.starting = true
	s.Mount()
	if err := s.Connect(); err != nil {
		return nil, err
	}
	res, err := s.Login(appName, envType, timeout)
	s.starting = false
	return res, err
}

func (s *SocketClientImpl) Shutdown() error {
	return s.Close(nil)
}
