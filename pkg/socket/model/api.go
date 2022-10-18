package model

type SocketClient interface {
	Mount()
	Send(frame Frame) error
	SendHeartbeat() error
	Close(err error) error
	Connect() error
	ReConnect(err error) error
	IsConnect() bool
	RequestTcp(path string, content interface{}, timeout int) (*Response, error)
}

type Implement interface {
	GetHeartbeatFrame() Frame
}

type CommonHeader struct {
	GlobalSeq string `json:"globalSeq" mapstructure:"globalSeq"`
	SubSeq    string `json:"subSeq" mapstructure:"subSeq"`
}

type RequestHeader struct {
	Path string `json:"path"`
	CommonHeader
}

type RequestDTO struct {
	Header RequestHeader `json:"header"`
	Body   interface{}   `json:"body"`
}

type Response struct {
	Code  int         `json:"code"`
	SeqId string      `json:"seqId"` // sequence number chosen by client
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

type ResponseDTO struct {
	Header CommonHeader `json:"header"`
	Body   Response     `json:"body"`
}
