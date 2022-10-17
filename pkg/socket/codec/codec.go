package codec

import (
	"bytes"
	"github.com/inooy/serco-client/pkg/socket/model"
)

type Codec interface {
	Decode(buffer *bytes.Buffer) ([]model.Frame, *bytes.Buffer)
	Encode(frame model.Frame) (*bytes.Buffer, error)
}
