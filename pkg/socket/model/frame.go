package model

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
)

type Command int8

const (
	CommandHeartbeat     Command = 2
	CommandLogin         Command = 3   // 登录请求
	CommandLoginResponse Command = 4   // 登录响应
	CommandAppPush       Command = 21  // APP推送
	CommandRequest       Command = 100 // 请求
	CommandReply         Command = 101 // 响应
	CommandEvent         Command = 102 // 发送事件
)

type Frame interface {
}

type PacketFrame struct {
	Cmd  Command
	Body interface{}
}

func (packet *PacketFrame) ToBuffer() (*bytes.Buffer, error) {
	body, err := json.Marshal(packet.Body)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(body), nil
}

type HeartbeatFrame struct {
	PacketFrame
}

var Heartbeat = HeartbeatFrame{
	PacketFrame: PacketFrame{
		Cmd: CommandHeartbeat,
	},
}

func (packet *HeartbeatFrame) ToBuffer() (*bytes.Buffer, error) {
	data := int8(-128)
	buffer := bytes.NewBuffer([]byte{})
	err := binary.Write(buffer, binary.BigEndian, data)
	if err != nil {
		return nil, err
	}
	return buffer, nil
}
