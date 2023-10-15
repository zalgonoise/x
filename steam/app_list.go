package steam

import (
	_ "embed"
	"encoding/json"
)

// fetch app_list manually and embed it in the binary, as it's not reasonable
// to download a 10MB file on each execution just for querying for an app ID
//
//go:embed internal/app_list/applist_*.json
var rawAppList []byte

func LoadAppsList() ([]App, error) {
	list := &AppsListHTTPResponse{}

	if err := json.Unmarshal(rawAppList, list); err != nil {
		return nil, err
	}

	return list.AppList.Apps.Apps, nil
}

type AppsListHTTPResponse struct {
	AppList AppListResponse `json:"applist"`
}

type AppListResponse struct {
	Apps AppsResponse `json:"apps"`
}

type AppsResponse struct {
	Apps []App `json:"app"`
}

type App struct {
	AppID int64  `json:"appid"`
	Name  string `json:"name"`
}
