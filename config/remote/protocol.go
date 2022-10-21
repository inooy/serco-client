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
1 魔数 v
2 协议版本号
3 cmd指令 数字 0~255
4 -7 Length 整个报文长度
8-Length body
*/

const HeadMetadataLen int = 7
const ProtocolVersion byte = 1
const Magic = byte('v')

type SercoCodec struct {
}

func deserializeFrame(cmd model.Command, buffer []byte) (model.Frame, error) {
	switch cmd {
	// 收到心跳结果
	case model.CommandHeartbeat:
		return model.Heartbeat, nil
	default:
		frame := model.PacketFrame{
			Cmd: cmd,
		}
		err := json.Unmarshal(buffer, &frame.Body)
		if err != nil {
			return nil, err
		}
		return frame, nil
	}
}

func (codec *SercoCodec) Decode(buffer []byte) ([]model.Frame, []byte) {
	var frames = make([]model.Frame, 0)
	offset := 0
	for {
		if len(buffer)-offset <= HeadMetadataLen {
			break
		}
		headBytes := buffer[offset:(offset + HeadMetadataLen)]
		headBuff := bytes.NewBuffer(headBytes)
		// read magic
		_, err := headBuff.ReadByte()
		if err != nil {
			panic(err)
		}
		// read version
		version, err := headBuff.ReadByte()

		if version != ProtocolVersion {
			panic(errors.New("protocol version not match: need 1, but real is " + strconv.Itoa(int(version))))
		}
		var cmd int8
		err = binary.Read(headBuff, binary.BigEndian, &cmd)
		var frameLength int32
		err = binary.Read(headBuff, binary.BigEndian, &frameLength)
		if err != nil {
			panic(err)
		}
		frameStart := offset + HeadMetadataLen
		frameEnd := offset + int(frameLength)
		if frameEnd > len(buffer) {
			// not all bytes of next frame received
			break
		}
		body := buffer[frameStart:frameEnd]
		frame, err := deserializeFrame(model.Command(cmd), body)
		if err != nil {
			panic(err)
		}
		frames = append(frames, frame)
		offset = frameEnd
	}
	return frames, buffer[offset:]
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

	var frameLength = framebuffer.Len()
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
	err = binary.Write(bf, binary.BigEndian, int32(frameLength+HeadMetadataLen))
	if err != nil {
		return
	}
	err = binary.Write(bf, binary.BigEndian, framebuffer.Bytes())
	if err != nil {
		return
	}
	return bf, nil
}
