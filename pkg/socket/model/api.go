package model

type SocketClient interface {
	Mount()
	Send(frame Frame) error
	SendHeartbeat() error
	Close(err error) error
	Connect() error
	ReConnect(err error) error
	IsConnect() bool
	RequestTcp(path string, content interface{}, timeout int) (interface{}, error)
}

type Implement interface {
	GetHeartbeatFrame() Frame
}
