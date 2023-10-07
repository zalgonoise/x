package steam

import (
	"bytes"

	"github.com/zalgonoise/x/errs"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/zalgonoise/x/steam/pb/proto/steam/store/v1"
)

const (
	headerMatcher = `{"app`
	wrapperHead   = `{"app_details":`
	wrapperTail   = `}`

	errDomain = "x/steam"

	ErrEmpty = errs.Kind("empty")

	ErrData = errs.Entity("data")
)

var (
	ErrEmptyData = errs.WithDomain(errDomain, ErrEmpty, ErrData)
)

// UnmarshalJSON decodes a JSON response of an HTTP GET call to the appdetails endpoint
// in the Steam store, as a store.AppDetailsResponse object.
//
// The underlying call is the following:
// GET https://store.steampowered.com/api/appdetails/?appids={comma_separated_app_ids}&cc={country_code}
func UnmarshalJSON(data []byte) (*store.AppDetailsResponse, error) {
	if len(data) == 0 {
		return nil, ErrEmptyData
	}

	response := &store.AppDetailsResponse{}

	if err := protojson.Unmarshal(addWrapper(data), response); err != nil {
		return nil, err
	}

	return response, nil
}

func addWrapper(data []byte) []byte {
	if len(data) < 5 {
		return data
	}

	if bytes.Equal(data[:5], []byte(headerMatcher)) {
		return data
	}

	buf := make([]byte, len(data)+len(wrapperHead)+len(wrapperTail))
	n := copy(buf, wrapperHead)
	n += copy(buf[n:], data)
	copy(buf[n:], wrapperTail)

	return buf
}
