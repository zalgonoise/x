package endpoints

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/zalgonoise/x/dns/store"
)

var (
	ErrInvalidBody = errors.New("invalid body")
	ErrInvalidJSON = errors.New("body contains invalid JSON")
	ErrInternal    = errors.New("internal error")
)

type StoreResponse struct {
	Success bool            `json:"success,omitempty"`
	Message string          `json:"message,omitempty"`
	Record  *store.Record   `json:"record,omitemtpy"`
	Records *[]store.Record `json:"records,omitemtpy"`
	Error   string          `json:"error,omitempty"`
}

func (e *endpoints) AddRecord(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		response, _ := json.Marshal(StoreResponse{
			Success: false,
			Message: ErrInvalidBody.Error(),
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}

	record := &store.Record{}
	err = json.Unmarshal(b, record)
	if err != nil {
		w.WriteHeader(400)
		response, _ := json.Marshal(StoreResponse{
			Success: false,
			Message: ErrInvalidJSON.Error(),
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}

	err = e.s.AddRecord(ctx, record)
	if err != nil {
		w.WriteHeader(500)
		response, _ := json.Marshal(StoreResponse{
			Success: false,
			Message: ErrInternal.Error(),
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}

	out, err := e.s.GetRecordByDomain(ctx, record)
	if err != nil {
		w.WriteHeader(500)
		response, _ := json.Marshal(StoreResponse{
			Success: true,
			Message: ErrInternal.Error(),
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}

	w.WriteHeader(200)
	response, _ := json.Marshal(StoreResponse{
		Success: true,
		Message: "added record successfully",
		Record:  out,
	})
	_, _ = w.Write(response)
}
func (e *endpoints) ListRecords(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	records, err := e.s.ListRecords(ctx)
	if err != nil {
		w.WriteHeader(500)
		response, _ := json.Marshal(StoreResponse{
			Success: false,
			Message: ErrInternal.Error(),
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}

	var out = make([]store.Record, len(records))
	for idx, record := range records {
		out[idx] = *record
	}

	w.WriteHeader(200)
	response, _ := json.Marshal(StoreResponse{
		Success: true,
		Message: "listing all records",
		Records: &out,
	})
	_, _ = w.Write(response)
}
func (e *endpoints) GetRecordByDomain(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		response, _ := json.Marshal(StoreResponse{
			Success: false,
			Message: ErrInvalidBody.Error(),
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}

	record := &store.Record{}
	err = json.Unmarshal(b, record)
	if err != nil {
		w.WriteHeader(400)
		response, _ := json.Marshal(StoreResponse{
			Success: false,
			Message: ErrInvalidJSON.Error(),
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}

	out, err := e.s.GetRecordByDomain(ctx, record)
	if err != nil {
		w.WriteHeader(500)
		response, _ := json.Marshal(StoreResponse{
			Success: true,
			Message: ErrInternal.Error(),
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}

	w.WriteHeader(200)
	response, _ := json.Marshal(StoreResponse{
		Success: true,
		Message: "fetched record for domain " + record.Name,
		Record:  out,
	})
	_, _ = w.Write(response)
}

func (e *endpoints) GetRecordByAddress(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		response, _ := json.Marshal(StoreResponse{
			Success: false,
			Message: ErrInvalidBody.Error(),
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}

	record := &store.Record{}
	err = json.Unmarshal(b, record)
	if err != nil {
		w.WriteHeader(400)
		response, _ := json.Marshal(StoreResponse{
			Success: false,
			Message: ErrInvalidJSON.Error(),
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}

	records, err := e.s.GetRecordByAddress(ctx, record.Addr)
	if err != nil {
		w.WriteHeader(500)
		response, _ := json.Marshal(StoreResponse{
			Success: true,
			Message: ErrInternal.Error(),
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}

	var out = make([]store.Record, len(records))
	for idx, record := range records {
		out[idx] = *record
	}

	w.WriteHeader(200)
	response, _ := json.Marshal(StoreResponse{
		Success: true,
		Message: "listing all records for IP address " + record.Addr,
		Records: &out,
	})
	_, _ = w.Write(response)
}
func (e *endpoints) UpdateRecord(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		response, _ := json.Marshal(StoreResponse{
			Success: false,
			Message: ErrInvalidBody.Error(),
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}

	record := &store.Record{}
	err = json.Unmarshal(b, record)
	if err != nil {
		w.WriteHeader(400)
		response, _ := json.Marshal(StoreResponse{
			Success: false,
			Message: ErrInvalidJSON.Error(),
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}

	err = e.s.UpdateRecord(ctx, record.Name, record)
	if err != nil {
		w.WriteHeader(500)
		response, _ := json.Marshal(StoreResponse{
			Success: true,
			Message: ErrInternal.Error(),
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}
	out, err := e.s.GetRecordByDomain(ctx, record)
	if err != nil {
		w.WriteHeader(500)
		response, _ := json.Marshal(StoreResponse{
			Success: true,
			Message: ErrInternal.Error(),
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}

	w.WriteHeader(200)
	response, _ := json.Marshal(StoreResponse{
		Success: true,
		Message: "updated record successfully",
		Record:  out,
	})
	_, _ = w.Write(response)
}
func (e *endpoints) DeleteRecord(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		response, _ := json.Marshal(StoreResponse{
			Success: false,
			Message: ErrInvalidBody.Error(),
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}

	record := &store.Record{}
	err = json.Unmarshal(b, record)
	if err != nil {
		w.WriteHeader(400)
		response, _ := json.Marshal(StoreResponse{
			Success: false,
			Message: ErrInvalidJSON.Error(),
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}

	err = e.s.DeleteRecord(ctx, record)
	if err != nil {
		w.WriteHeader(500)
		response, _ := json.Marshal(StoreResponse{
			Success: true,
			Message: ErrInternal.Error(),
			Error:   err.Error(),
		})
		_, _ = w.Write(response)
		return
	}

	w.WriteHeader(200)
	response, _ := json.Marshal(StoreResponse{
		Success: true,
		Message: "record deleted successfully",
	})
	_, _ = w.Write(response)
}
