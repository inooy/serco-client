package codec

import (
	"bytes"
	"github.com/inooy/serco-client/pkg/socket/model"
)

type Codec interface {
	Decode(buffer []byte) ([]model.Frame, []byte)
	Encode(frame model.Frame) (*bytes.Buffer, error)
}
