package client

import (
	"context"
	"fmt"
	"github.com/inooy/serco-client/pkg/log"
	"github.com/inooy/serco-client/pkg/socket/codec"
	"github.com/inooy/serco-client/pkg/socket/connection"
	"github.com/inooy/serco-client/pkg/socket/model"
	"github.com/mitchellh/mapstructure"
	"time"
)

var NooyCodec = codec.NooyCodec{}

func RequestTcp(addr string, path string, content interface{}, timeout int) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	var result interface{}
	var err error

	conn := connection.NewTcpConnection(connection.TcpSocketConnectOpts{
		Host: addr,
	}, &NooyCodec)

	go func(ctx context.Context) {
		// 发送HTTP请求
		conn.AddListener(connection.Listener{
			OnReceive: func(frame model.Frame) {
				packet := (frame).(model.PacketFrame)
				var response model.ResponseDTO
				//将 map 转换为指定的结构体
				if err := mapstructure.Decode(packet.Body, &response); err != nil {
					fmt.Println(err)
				}
				result = response.Body
				cancel()
			},
			OnError: func(err error) {
				cancel()
			},
		})

		err = conn.Connect()
		if err != nil {
			log.Error("连接失败")
			cancel()
		}
		requestDTO := model.RequestDTO{
			Body: content,
		}
		requestDTO.Header = model.RequestHeader{
			Path: path,
		}
		packet := model.PacketFrame{Cmd: 100, Body: requestDTO}
		err = conn.Send(packet)
		if err != nil {
			log.Errorf("发送失败: %s", err.Error())
			cancel()
		}
	}(ctx)

	defer func(conn connection.SocketConnection) {
		log.Info("处理中断")
		err := conn.Close(nil)
		if err != nil {
			log.Error("关闭失败", err.Error())
		}
	}(conn)

	select {
	case <-ctx.Done():
		fmt.Println("call successfully!!!")
		return result, err
	}
}
