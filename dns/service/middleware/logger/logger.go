package logger

import (
	"context"
	"time"

	dnsr "github.com/miekg/dns"

	"github.com/zalgonoise/x/dns/health"
	"github.com/zalgonoise/x/dns/service"
	"github.com/zalgonoise/x/dns/store"
	"github.com/zalgonoise/zlog/log"
	"github.com/zalgonoise/zlog/log/event"
)

type LoggedService struct {
	svc    service.Service
	logger log.Logger
}

func LogService(svc service.Service, logger log.Logger) service.Service {
	return &LoggedService{
		svc:    svc,
		logger: logger,
	}
}

func (s *LoggedService) AddRecord(ctx context.Context, r *store.Record) error {
	s.logger.Log(event.New().
		Level(event.Level_debug).
		Prefix("service").
		Sub("records").
		Message("AddRecord request").
		Metadata(event.Field{
			"input": r,
		}).
		Build())

	err := s.svc.AddRecord(ctx, r)
	if err != nil {
		s.logger.Log(event.New().
			Level(event.Level_warn).
			Prefix("service").
			Sub("records").
			Message("AddRecord error").
			Metadata(event.Field{
				"error": err.Error(),
			}).
			Build())
		return err
	}
	s.logger.Log(event.New().
		Level(event.Level_debug).
		Prefix("service").
		Sub("records").
		Message("AddRecord success").
		Build())
	return nil
}
func (s *LoggedService) AddRecords(ctx context.Context, rs ...*store.Record) error {
	s.logger.Log(event.New().
		Level(event.Level_debug).
		Prefix("service").
		Sub("records").
		Message("AddRecords request").
		Metadata(event.Field{
			"input": rs,
			"len":   len(rs),
		}).
		Build())

	err := s.svc.AddRecords(ctx, rs...)
	if err != nil {
		s.logger.Log(event.New().
			Level(event.Level_warn).
			Prefix("service").
			Sub("records").
			Message("AddRecords error").
			Metadata(event.Field{
				"error": err.Error(),
			}).
			Build())
		return err
	}
	s.logger.Log(event.New().
		Level(event.Level_debug).
		Prefix("service").
		Sub("records").
		Message("AddRecords success").
		Build())
	return nil
}
func (s *LoggedService) ListRecords(ctx context.Context) ([]*store.Record, error) {
	s.logger.Log(event.New().
		Level(event.Level_debug).
		Prefix("service").
		Sub("records").
		Message("ListRecords request").
		Build())

	rs, err := s.svc.ListRecords(ctx)
	if err != nil {
		s.logger.Log(event.New().
			Level(event.Level_warn).
			Prefix("service").
			Sub("records").
			Message("ListRecords error").
			Metadata(event.Field{
				"error": err.Error(),
			}).
			Build())
		return rs, err
	}
	s.logger.Log(event.New().
		Level(event.Level_debug).
		Prefix("service").
		Sub("records").
		Message("ListRecords success").
		Metadata(event.Field{
			"output": rs,
			"len":    len(rs),
		}).
		Build())
	return rs, nil
}
func (s *LoggedService) GetRecordByDomain(ctx context.Context, r *store.Record) (*store.Record, error) {
	s.logger.Log(event.New().
		Level(event.Level_debug).
		Prefix("service").
		Sub("records").
		Message("GetRecordByDomain request").
		Metadata(event.Field{
			"input": r,
		}).
		Build())

	out, err := s.svc.GetRecordByDomain(ctx, r)
	if err != nil {
		s.logger.Log(event.New().
			Level(event.Level_warn).
			Prefix("service").
			Sub("records").
			Message("GetRecordByDomain error").
			Metadata(event.Field{
				"error": err.Error(),
			}).
			Build())
		return out, err
	}
	s.logger.Log(event.New().
		Level(event.Level_debug).
		Prefix("service").
		Sub("records").
		Message("GetRecordByDomain success").
		Metadata(event.Field{
			"output": out,
		}).
		Build())
	return out, nil
}
func (s *LoggedService) GetRecordByAddress(ctx context.Context, address string) ([]*store.Record, error) {
	s.logger.Log(event.New().
		Level(event.Level_debug).
		Prefix("service").
		Sub("records").
		Message("GetRecordByAddress request").
		Metadata(event.Field{
			"input": address,
		}).
		Build())

	rs, err := s.svc.GetRecordByAddress(ctx, address)
	if err != nil {
		s.logger.Log(event.New().
			Level(event.Level_warn).
			Prefix("service").
			Sub("records").
			Message("GetRecordByAddress error").
			Metadata(event.Field{
				"error": err.Error(),
			}).
			Build())
		return rs, err
	}
	s.logger.Log(event.New().
		Level(event.Level_debug).
		Prefix("service").
		Sub("records").
		Message("GetRecordByAddress success").
		Metadata(event.Field{
			"output": rs,
			"len":    len(rs),
		}).
		Build())
	return rs, nil
}
func (s *LoggedService) UpdateRecord(ctx context.Context, domain string, r *store.Record) error {
	s.logger.Log(event.New().
		Level(event.Level_debug).
		Prefix("service").
		Sub("records").
		Message("UpdateRecord request").
		Metadata(event.Field{
			"target": domain,
			"input":  r,
		}).
		Build())

	err := s.svc.UpdateRecord(ctx, domain, r)
	if err != nil {
		s.logger.Log(event.New().
			Level(event.Level_warn).
			Prefix("service").
			Sub("records").
			Message("UpdateRecord error").
			Metadata(event.Field{
				"error": err.Error(),
			}).
			Build())
		return err
	}
	s.logger.Log(event.New().
		Level(event.Level_debug).
		Prefix("service").
		Sub("records").
		Message("UpdateRecord success").
		Build())
	return nil
}
func (s *LoggedService) DeleteRecord(ctx context.Context, r *store.Record) error {
	s.logger.Log(event.New().
		Level(event.Level_debug).
		Prefix("service").
		Sub("records").
		Message("DeleteRecord request").
		Metadata(event.Field{
			"input": r,
		}).
		Build())

	err := s.svc.DeleteRecord(ctx, r)
	if err != nil {
		s.logger.Log(event.New().
			Level(event.Level_warn).
			Prefix("service").
			Sub("records").
			Message("DeleteRecord error").
			Metadata(event.Field{
				"error": err.Error(),
			}).
			Build())
		return err
	}
	s.logger.Log(event.New().
		Level(event.Level_debug).
		Prefix("service").
		Sub("records").
		Message("DeleteRecord success").
		Build())
	return nil
}
func (s *LoggedService) AnswerDNS(r *store.Record, m *dnsr.Msg) {
	s.logger.Log(event.New().
		Level(event.Level_debug).
		Prefix("service").
		Sub("dns").
		Message("AnswerDNS request").
		Build())

	s.svc.AnswerDNS(r, m)
	go func() {
		time.Sleep(5 * time.Millisecond)

		s.logger.Log(event.New().
			Level(event.Level_debug).
			Prefix("service").
			Sub("dns").
			Message("AnswerDNS response").
			Metadata(event.Field{
				"output": r,
			}).
			Build())
	}()
}

func (s *LoggedService) StoreHealth(entries int, t time.Duration) *health.Report {
	s.logger.Log(event.New().
		Level(event.Level_debug).
		Prefix("service").
		Sub("health").
		Message("StoreHealth request").
		Build())

	r := s.svc.StoreHealth(entries, t)
	go func() {
		time.Sleep(5 * time.Millisecond)

		s.logger.Log(event.New().
			Level(event.Level_debug).
			Prefix("service").
			Sub("health").
			Message("StoreHealth response").
			Metadata(event.Field{
				"output": r,
			}).
			Build())
	}()
	return r
}

func (s *LoggedService) DNSHealth() *health.Report {
	s.logger.Log(event.New().
		Level(event.Level_debug).
		Prefix("service").
		Sub("health").
		Message("DNSHealth request").
		Build())

	r := s.svc.DNSHealth()
	go func() {
		time.Sleep(5 * time.Millisecond)

		s.logger.Log(event.New().
			Level(event.Level_debug).
			Prefix("service").
			Sub("health").
			Message("DNSHealth response").
			Metadata(event.Field{
				"output": r,
			}).
			Build())
	}()
	return r
}
