package steam

import (
	"bytes"
	"encoding/json"

	"github.com/zalgonoise/x/errs"
	pb "github.com/zalgonoise/x/steam/pb/proto/steam/store/v1"
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

type Result[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
}

type FilterResult[T any] struct {
	Success bool         `json:"success"`
	Data    map[string]T `json:"data"`
}

func Get[T any](data []byte, filter string) (map[string]*T, error) {
	if filter == "" {
		return getAll[T](data)
	}

	return getFilter[T](data, filter)
}

// UnmarshalJSON decodes a JSON response of an HTTP GET call to the appdetails endpoint
// in the Steam store, as a map[string]pb.Data, linking appids to pb.Data objects.
//
// The underlying call is the following:
// GET https://store.steampowered.com/api/appdetails/?appids={comma_separated_app_ids}&cc={country_code}
func UnmarshalJSON(data []byte) (map[string]*pb.Data, error) {
	if len(data) == 0 {
		return nil, ErrEmptyData
	}

	return Get[pb.Data](data, "")
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

func getAll[T any](data []byte) (map[string]*T, error) {
	response := &map[string]Result[T]{}

	if err := json.Unmarshal(data, response); err != nil {
		return nil, err
	}

	output := make(map[string]*T, len(*response))

	for id, priceData := range *response {
		res := new(T)

		buf, err := json.Marshal(priceData)
		if err != nil {
			return nil, err
		}

		if err = json.Unmarshal(buf, res); err != nil {
			return nil, err
		}

		output[id] = res
	}

	return output, nil
}

func getFilter[T any](data []byte, filter string) (map[string]*T, error) {
	response := &map[string]FilterResult[T]{}

	if err := json.Unmarshal(data, response); err != nil {
		return nil, err
	}

	output := make(map[string]*T, len(*response))

	for id, filteredDataRaw := range *response {
		if filteredData, ok := filteredDataRaw.Data[filter]; ok {
			res := new(T)

			buf, err := json.Marshal(filteredData)
			if err != nil {
				return nil, err
			}

			if err = json.Unmarshal(buf, res); err != nil {
				return nil, err
			}

			output[id] = res
		}
	}

	return output, nil
}
