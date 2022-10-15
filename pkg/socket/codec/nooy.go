package codec

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"github.com/inooy/serco-client/pkg/socket/model"
)

/**
协议说明：
字节编码：
1 协议版本号
2 cmd指令 数字 0~255
3 mask
4 -7 Length 整个报文长度
8-Length body
*/

const HeadMetadataLen int32 = 7
const ProtocolVersion int8 = 1

type NooyCodec struct {
}

func deserializeFrame(cmd model.Command, buffer bytes.Buffer) (model.Frame, error) {
	switch cmd {
	// 收到心跳结果
	case model.COMMAND_HEARTBEAT_REQ:
		return model.Heartbeat, nil
	default:
		frame := model.PacketFrame{
			Cmd: cmd,
		}
		err := json.Unmarshal(buffer.Bytes(), &frame.Body)
		if err != nil {
			return nil, err
		}
		return frame, nil
	}
}

func (codec *NooyCodec) Decode(buffer *bytes.Buffer) ([]model.Frame, *bytes.Buffer) {
	var frames = make([]model.Frame, 0)
	for {
		if HeadMetadataLen >= int32(buffer.Len()) {
			break
		}
		//read version
		_, err := buffer.ReadByte()
		if err != nil {
			panic(err)
		}
		//read mask
		_, err = buffer.ReadByte()
		var cmd int8
		err = binary.Read(buffer, binary.BigEndian, &cmd)
		var frameLength int32
		err = binary.Read(buffer, binary.BigEndian, &frameLength)
		if err != nil {
			panic(err)
		}
		if frameLength > int32(buffer.Len()) {
			// not all bytes of next frame received
			break
		}
		body := make([]byte, frameLength)
		_, err = buffer.Read(body)
		if err != nil {
			panic(err)
		}
		frameBuffer := bytes.NewBuffer(body)
		frame, err := deserializeFrame(model.Command(cmd), *frameBuffer)
		if err != nil {
			panic(err)
		}
		frames = append(frames, frame)
	}
	return frames, buffer
}

func (codec *NooyCodec) Encode(frame model.Frame) *bytes.Buffer {
	packet := (frame).(model.PacketFrame)
	framebuffer, err := packet.ToBuffer()
	if err != nil {
		panic(err)
	}
	var frameLength = int32(framebuffer.Len())
	mask := ProtocolVersion | 0b01111111
	mask = mask | 64
	mask = mask & (32 ^ 0b01111111)
	mask = mask | 16
	var cmdByte = int8(0x00 | packet.Cmd)

	// 先组包好，一次性写入IO，一定程度避免半包粘包，网络IO比内存读写更耗时
	var bf bytes.Buffer
	err = binary.Write(&bf, binary.BigEndian, ProtocolVersion)
	if err != nil {
		panic(err)
	}
	err = binary.Write(&bf, binary.BigEndian, mask)
	if err != nil {
		panic(err)
	}
	err = binary.Write(&bf, binary.BigEndian, cmdByte)
	if err != nil {
		panic(err)
	}
	err = binary.Write(&bf, binary.BigEndian, frameLength)
	if err != nil {
		panic(err)
	}
	err = binary.Write(&bf, binary.BigEndian, framebuffer.Bytes())
	if err != nil {
		panic(err)
	}
	return &bf
}
