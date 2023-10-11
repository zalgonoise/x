package steam

import (
	"encoding/json"

	"github.com/zalgonoise/x/errs"
)

const (
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

// Get decodes a JSON response of an HTTP GET call to the appdetails endpoint.
//
// If the filter string is unset, it returns a map[string]*pb.Data, linking appids
// to pb.Data objects.
//
// If a filter is provided, the returned value will be a map[string]*T.
//
// The underlying call is the following:
// GET https://store.steampowered.com/api/appdetails/?appids={comma_separated_app_ids}&cc={country_code}
func Get[T any](data []byte, filter string) (map[string]*T, error) {
	if len(data) == 0 {
		return nil, ErrEmptyData
	}

	if filter == "" {
		return getAll[T](data)
	}

	return getFilter[T](data, filter)
}

func getAll[T any](data []byte) (map[string]*T, error) {
	response := &map[string]Result[*T]{}

	if err := json.Unmarshal(data, response); err != nil {
		return nil, err
	}

	output := make(map[string]*T, len(*response))

	for id, objData := range *response {
		output[id] = objData.Data
	}

	return output, nil
}

func getFilter[T any](data []byte, filter string) (map[string]*T, error) {
	response := &map[string]FilterResult[*T]{}

	if err := json.Unmarshal(data, response); err != nil {
		// some objects are not encapsulated like the others. If it fails to unmarshal into a map,
		// try to unmarshal into an object of the input type.
		return getAll[T](data)
	}

	output := make(map[string]*T, len(*response))

	for id, filteredDataRaw := range *response {
		if filteredData, ok := filteredDataRaw.Data[filter]; ok {
			output[id] = filteredData
		}
	}

	return output, nil
}
