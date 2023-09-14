package httplog

import "encoding/json"

type MarshalJSON struct {
	indent bool
}

func (m MarshalJSON) Encode(r HTTPRecord) ([]byte, error) {
	if m.indent {
		return json.MarshalIndent(r, "", "  ")
	}

	return json.Marshal(r)
}
