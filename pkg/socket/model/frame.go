package model

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
)

type Command int8

const (
	COMMAND_HEARTBEAT_REQ Command = 13
	COMMAND_UNKNOWN       Command = 0
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
	Cmd: COMMAND_HEARTBEAT_REQ,
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
