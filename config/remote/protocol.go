package remote

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/inooy/serco-client/pkg/socket/model"
	"strconv"
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

var VersionContent = int8(1)
var Magic = byte('v')

type SercoCodec struct {
}

func deserializeFrame(cmd model.Command, buffer bytes.Buffer) (model.Frame, error) {
	switch cmd {
	// 收到心跳结果
	case model.CommandHeartbeat:
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

func (codec *SercoCodec) Decode(buffer *bytes.Buffer) ([]model.Frame, *bytes.Buffer) {
	var frames = make([]model.Frame, 0)
	for {
		if HeadMetadataLen >= int32(buffer.Len()) {
			break
		}
		// read magic
		_, err := buffer.ReadByte()
		if err != nil {
			panic(err)
		}
		// read version
		version, err := buffer.ReadByte()
		if version != 1 {
			panic(errors.New("protocol version not match: need 1, but real is " + strconv.Itoa(int(version))))
		}
		var cmd int8
		err = binary.Read(buffer, binary.BigEndian, &cmd)
		var frameLength int32
		err = binary.Read(buffer, binary.BigEndian, &frameLength)
		if err != nil {
			panic(err)
		}
		if frameLength-HeadMetadataLen > int32(buffer.Len()) {
			// not all bytes of next frame received
			break
		}
		body := make([]byte, frameLength-HeadMetadataLen)
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

func (codec *SercoCodec) Encode(frame model.Frame) (bf *bytes.Buffer, err error) {
	var cmd model.Command
	var framebuffer *bytes.Buffer
	bf = new(bytes.Buffer)

	switch frame.(type) {
	case *model.HeartbeatFrame:
		cmd = model.Heartbeat.Cmd
		buf, err := model.Heartbeat.ToBuffer()
		if err != nil {
			return bf, err
		}
		framebuffer = buf
	case model.PacketFrame:
		packet := (frame).(model.PacketFrame)
		cmd = packet.Cmd
		buf, err := packet.ToBuffer()
		if err != nil {
			return bf, err
		}
		framebuffer = buf
	}

	var frameLength = int32(framebuffer.Len())
	//mask := ProtocolVersion | 0b01111111
	//mask = mask | 64
	//mask = mask & (32 ^ 0b01111111)
	//mask = mask | 16
	var cmdByte = int8(0x00 | cmd)

	// 先组包好，一次性写入IO，一定程度避免半包粘包，网络IO比内存读写更耗时
	err = binary.Write(bf, binary.BigEndian, Magic)
	if err != nil {
		return
	}
	err = binary.Write(bf, binary.BigEndian, ProtocolVersion)
	if err != nil {
		return
	}
	err = binary.Write(bf, binary.BigEndian, cmdByte)
	if err != nil {
		return
	}
	err = binary.Write(bf, binary.BigEndian, frameLength+HeadMetadataLen)
	if err != nil {
		return
	}
	err = binary.Write(bf, binary.BigEndian, framebuffer.Bytes())
	if err != nil {
		return
	}
	return bf, nil
}
