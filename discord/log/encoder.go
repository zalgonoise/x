package log

import (
	"encoding/json"
	"strings"

	"github.com/switchupcb/dasgo/v10/dasgo"
	"github.com/zalgonoise/x/pluslog/httplog"
)

type Encoder struct {
	e httplog.Encoder
}

type MarshalJSON struct {
	indent bool
}

func (m MarshalJSON) Encode(r httplog.HTTPRecord) ([]byte, error) {
	if m.indent {
		return json.MarshalIndent(r, "", "  ")
	}

	return json.Marshal(r)
}

func New(e httplog.Encoder) httplog.Encoder {
	return Encoder{
		e: e,
	}
}

func JSON(indent bool) httplog.Encoder {
	return Encoder{
		e: MarshalJSON{
			indent: indent,
		},
	}
}

func (e Encoder) Encode(record httplog.HTTPRecord) ([]byte, error) {
	data, err := e.e.Encode(record)
	if err != nil {
		return nil, err
	}

	sb := new(strings.Builder)
	sb.WriteString("```")
	sb.Write(data)
	sb.WriteString("```")

	out := sb.String()

	content := &dasgo.ExecuteWebhook{
		Content: &out,
	}

	buf, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
