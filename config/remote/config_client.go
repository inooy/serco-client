package remote

import (
	"fmt"
	"github.com/inooy/serco-client/pkg/log"
	"github.com/inooy/serco-client/pkg/socket/client"
	"github.com/inooy/serco-client/pkg/socket/codec"
	"github.com/inooy/serco-client/pkg/socket/connection"
	"github.com/inooy/serco-client/pkg/socket/model"
	"github.com/mitchellh/mapstructure"
)

type SocketClientImpl struct {
	*client.Template

	heartBeatFrame *model.HeartbeatFrame
}

func NewConfigSocketClient(options connection.TcpSocketConnectOpts) client.SocketClient {
	impl := &SocketClientImpl{}
	conn := connection.NewTcpConnection(options, &codec.NooyCodec{})
	template := client.NewTemplate(impl, conn)
	impl.Template = template
	conn.AddListener(connection.Listener{
		OnReceive: func(frame model.Frame) {
			impl.handleFrame(frame)
		},
	})
	return impl
}

func (s *SocketClientImpl) handleFrame(frame model.Frame) {
	packet := frame.(model.PacketFrame)
	if packet.Cmd == model.CommandHeartbeat {
		s.ReceiveHeartbeat()
		return
	} else if packet.Cmd == model.CommandAppPush {
		// todo 收到配置推送更新，处理配置更新
		log.Info("receive app push")
	} else if packet.Cmd == model.CommandReply {
		var response client.ResponseDTO
		//将 map 转换为指定的结构体
		if err := mapstructure.Decode(packet.Body, &response); err != nil {
			fmt.Println(err)
		}
		// 发布事件
		s.Emitter.Emit(response.Header.SubSeq, &response)
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

func (s *SocketClientImpl) Launch() error {
	return s.Connect()
}

func (s *SocketClientImpl) Shutdown() error {
	return s.Close(nil)
}
