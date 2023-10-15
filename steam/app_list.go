package steam

import (
	"context"
	_ "embed"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	appsListURL  = "https://api.steampowered.com/ISteamApps/GetAppList/v0001/"
	baseFilePath = "internal/app_list/applist.json"
)

// fetch app_list manually and embed it in the binary, as it's not reasonable
// to download a 10MB file on each execution just for querying for an app ID
//
//go:embed internal/app_list/applist.json
var rawAppList []byte

func GetAppsList() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, appsListURL, http.NoBody)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")

	res, err := (&http.Client{
		Transport: http.DefaultTransport,
		Timeout:   time.Minute,
	}).Do(req)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	f, err := os.Create(baseFilePath)

	if err != nil {
		return err
	}

	if _, err = io.Copy(f, res.Body); err != nil {
		return err
	}

	return nil
}

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
