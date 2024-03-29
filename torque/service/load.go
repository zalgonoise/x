package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/zalgonoise/x/torque/vehicles"
)

const (
	vehiclesDataURI         = "https://raw.githubusercontent.com/DurtyFree/gta-v-data-dumps/master/vehicles.json"
	vehiclesHandlingDataURI = "https://raw.githubusercontent.com/DurtyFree/gta-v-data-dumps/master/vehicleHandlings.json"
)

func (s *Service) Load() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if err := s.loadVehicles(ctx); err != nil {
		return err
	}

	if err := s.loadHandling(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Service) loadVehicles(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, vehiclesDataURI, http.NoBody)
	if err != nil {
		return err
	}

	res, err := (&http.Client{
		Transport: http.DefaultTransport,
		Timeout:   time.Minute,
	}).Do(req)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	v := make([]vehicles.Vehicle, 0, 1024)

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	if err = s.repo.BulkInsertVehicles(ctx, v); err != nil {
		return err
	}

	return nil
}

func (s *Service) loadHandling(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, vehiclesHandlingDataURI, http.NoBody)
	if err != nil {
		return err
	}

	res, err := (&http.Client{
		Transport: http.DefaultTransport,
		Timeout:   time.Minute,
	}).Do(req)

	if err != nil {
		return err
	}

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	v := make([]vehicles.Handling, 0, 1024)

	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	if err = s.repo.BulkInsertHandling(ctx, v); err != nil {
		return err
	}

	return nil
}
