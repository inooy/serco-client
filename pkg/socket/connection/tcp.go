package connection

import (
	"github.com/inooy/serco-client/pkg/log"
	"github.com/inooy/serco-client/pkg/socket/codec"
	"net"
)

type TcpOpts struct {
	Host string `json:"host"`
}

type TcpConnection struct {
	*template
	options TcpOpts
}

func NewTcpConnection(options TcpOpts, codecImpl codec.Codec) SocketConnection {
	impl := &TcpConnection{
		options: options,
	}
	template := newTemplate(impl, codecImpl)
	impl.template = template
	return impl
}

func (conn *TcpConnection) Connect() error {
	if conn.GetConnectionState() == CONNECTED {
		log.Warn("tcp can not connect: a connection is already established.")
		return nil
	} else if conn.GetConnectionState() == CONNECTING {
		log.Warn("tcp can not connect: a connection is connecting.")
		return nil
	}
	// tcp
	conn.SetConnectionStatus(CONNECTING)
	socket, err := net.Dial("tcp", conn.options.Host)
	if err != nil {
		conn.SetConnectionStatus(ERROR)
		return err
	}
	conn.Conn = socket
	go conn.setupSocket(socket)
	conn.SetConnectionStatus(CONNECTED)
	return nil
}
