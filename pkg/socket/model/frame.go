package model

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
)

type Command int8

const (
	CommandHeartbeat Command = 13
	CommandAppPush           = 21  // APP推送
	CommandRequest           = 100 // 请求
	CommandReply             = 101 // 响应
	CommandEvent             = 102 // 发送事件
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
	Cmd Command
}

var Heartbeat = HeartbeatFrame{
	Cmd: CommandHeartbeat,
}

func (packet *HeartbeatFrame) ToBuffer() (*bytes.Buffer, error) {
	data := int32(-128)
	buffer := bytes.NewBuffer([]byte{})
	err := binary.Write(buffer, binary.BigEndian, data)
	if err != nil {
		return nil, err
	}
	return buffer, nil
}
