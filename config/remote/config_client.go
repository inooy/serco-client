package remote

import (
	"errors"
	"fmt"
	"github.com/inooy/serco-client/pkg/log"
	"github.com/inooy/serco-client/pkg/socket/client"
	"github.com/inooy/serco-client/pkg/socket/connection"
	"github.com/inooy/serco-client/pkg/socket/model"
	"github.com/inooy/serco-client/pkg/tools"
	"github.com/mitchellh/mapstructure"
	"time"
)

type LoginRequest struct {
	SeqId   string `json:"seqId"` // sequence number chosen by client
	Token   string `json:"token"`
	AppName string `json:"appName"`
	EnvType string `json:"envType"`
}

type CheckRequest struct {
	AppName string         `json:"appName" mapstructure:"appName"`
	EnvType string         `json:"envType" mapstructure:"envType"`
	Old     map[string]int `json:"old" mapstructure:"old"`
}

type AppPushMsg struct {
	MsgId    string   `json:"msgId" mapstructure:"msgId"` // 推送消息id
	AppName  string   `json:"appName" mapstructure:"appName"`
	EnvType  string   `json:"envType" mapstructure:"envType"`
	Metadata Metadata `json:"metadata" mapstructure:"metadata"`
}

type SocketClientImpl struct {
	*client.Template
	listener       ConfigFileListener
	loginCallback  func(*model.Response)
	heartBeatFrame *model.HeartbeatFrame
	starting       bool
	appName        string
	envType        string
}

func NewConfigSocketClient(options connection.TcpSocketConnectOpts, listener ConfigFileListener) *SocketClientImpl {
	impl := &SocketClientImpl{
		listener: listener,
	}
	conn := connection.NewTcpConnection(options, &SercoCodec{})
	template := client.NewTemplate(impl, conn)
	impl.Template = template
	conn.AddListener(connection.Listener{
		OnReceive: func(frame model.Frame) {
			impl.handleFrame(frame)
		},
		OnStatusChange: func(status connection.Status) {
			// 非启动时重连，那么重新登录
			if status == connection.CONNECTED && !impl.starting {
				log.Info("reconnect success, start re login...")
				result, err := impl.Login(impl.appName, impl.envType, 6000)
				if err != nil {
					_ = impl.Close(err)
					return
				}
				if result.Code != 200 {
					_ = impl.Close(err)
					return
				}
				log.Info("re login success!")
				var list []Metadata
				//将 map 转换为指定的结构体
				if err = mapstructure.Decode(result.Data, &list); err != nil {
					_ = impl.Close(err)
					return
				}
				for i := range list {
					impl.listener.OnFileChange(&list[i])
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
	if packet.Cmd == model.CommandAppPush {
		// todo 收到配置推送更新，处理配置更新
		log.Info("receive app push")
		var msg AppPushMsg
		//将 map 转换为指定的结构体
		if err := mapstructure.Decode(packet.Body, &msg); err != nil {
			fmt.Println(err)
		}
		s.listener.OnFileChange(&msg.Metadata)
	} else if packet.Cmd == model.CommandReply {
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
	s.appName = appName
	s.envType = envType
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
