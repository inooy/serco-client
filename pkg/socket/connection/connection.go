package connection

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/inooy/serco-client/pkg/socket/codec"
	"github.com/inooy/serco-client/pkg/socket/model"
	"io"
	"net"
)

type SocketConnection interface {
	AddListener(listener Listener)
	GetConnectionState() Status
	SetConnectionStatus(status Status)
	Close(ex error) (err error)
	Send(frame model.Frame) error
	implement
}

type implement interface {
	Connect() error
}

type template struct {
	implement
	Conn      net.Conn
	Buf       bytes.Buffer
	Codec     codec.Codec
	status    Status
	listeners []Listener
}

func newTemplate(impl implement, code codec.Codec) *template {
	return &template{
		implement: impl,
		Codec:     code,
		status:    NotConnected,
	}
}

type Status string

const (
	CLOSED       Status = "CLOSED"
	CONNECTED    Status = "CONNECTED"
	CONNECTING   Status = "CONNECTING"
	NotConnected Status = "NOT_CONNECTED"
	ERROR        Status = "ERROR"
)

type Listener struct {
	OnStatusChange func(status Status)
	OnReceive      func(frame model.Frame)
	OnError        func(err error)
}

func (conn *template) AddListener(listener Listener) {
	conn.listeners = append(conn.listeners, listener)
}

func (conn *template) GetConnectionState() Status {
	return conn.status
}

func (conn *template) SetConnectionStatus(status Status) {
	// EVENT
	conn.status = status
	for i := range conn.listeners {
		if conn.listeners[i].OnStatusChange != nil {
			conn.listeners[i].OnStatusChange(status)
		}
	}
}

func (conn *template) Close(ex error) (err error) {
	if ex != nil {
		conn.handleError(ex)
		return
	}
	return conn.close(nil)
}

func (conn *template) close(ex error) (err error) {
	if conn.status == CLOSED || conn.status == ERROR {
		// already closed
		return
	}
	if ex == nil {
		conn.SetConnectionStatus(ERROR)
	} else {
		conn.SetConnectionStatus(CLOSED)
	}
	if conn.Conn != nil {
		err = conn.Conn.Close()
		conn.Conn = nil
	}
	return err
}

func (conn *template) Send(frame model.Frame) error {
	if conn.Conn == nil {
		return errors.New("SocketConnection: Cannot send frame, not connected")
	}
	_buffer := conn.Codec.Encode(frame)
	_, err := conn.Conn.Write(_buffer.Bytes())
	if err != nil {
		conn.handleError(err)
		return err
	}
	return nil
}

func (conn *template) handleData(buffer []byte) {
	frames := conn.readFrames(buffer)
	for _, frame := range frames {
		for i := range conn.listeners {
			if conn.listeners[i].OnReceive != nil {
				conn.listeners[i].OnReceive(frame)
			}
		}
	}
}

func (conn *template) handleError(err error) {
	if err == nil {
		err = errors.New("socketConnection: Socket closed unexpectedly")
	}
	for i := range conn.listeners {
		if conn.listeners[i].OnError != nil {
			conn.listeners[i].OnError(err)
		}
	}
	_ = conn.close(err)
}

func (conn *template) readFrames(buffer []byte) []model.Frame {
	conn.Buf.Write(buffer)
	frames, remaining := conn.Codec.Decode(&conn.Buf)
	conn.Buf = *remaining
	return frames
}

func (conn *template) setupSocket(real net.Conn) {
	buffer := make([]byte, 1024)
	for {
		readLen, err := real.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("read error: " + err.Error())
			conn.handleError(err)
			return
		}
		conn.handleData(buffer[:readLen])
	}
}
