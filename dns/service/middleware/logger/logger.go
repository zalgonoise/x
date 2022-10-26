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

// LoggedService will wrap a service.Service with a logger and
// register the incoming events, as well as their outcome
type LoggedService struct {
	svc    service.Service
	logger log.Logger
}

// LogService will return a LoggedService in the form of a service.Service,
// by wraping an input service.Service `svc` with log.Logger `logger`
func LogService(svc service.Service, logger log.Logger) service.Service {
	return &LoggedService{
		svc:    svc,
		logger: logger,
	}
}

// AddRecord uses the store.Repository to create a DNS Record
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

// AddRecords uses the store.Repository to create a set of DNS Records
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

// ListRecord uses the store.Repository to return all DNS Records
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

// GetRecordByDomain uses the store.Repository to return the DNS Record associated with
// the domain name and record type found in store.Record `r`
func (s *LoggedService) GetRecordByTypeAndDomain(ctx context.Context, rtype, domain string) (*store.Record, error) {
	s.logger.Log(event.New().
		Level(event.Level_debug).
		Prefix("service").
		Sub("records").
		Message("GetRecordByDomain request").
		Metadata(event.Field{
			"input": map[string]string{
				"type":   rtype,
				"domain": domain,
			},
		}).
		Build())

	out, err := s.svc.GetRecordByTypeAndDomain(ctx, rtype, domain)
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

// GetRecordByDomain uses the store.Repository to return the DNS Records associated with
// the IP address found in store.Record `r`
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

// UpdateRecord uses the store.Repository to update the record with domain name `domain`,
// based on the data provided in store.Record `r`
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

// DeleteRecord uses the store.Repository to remove the store.Record based on input `r`
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

// AnswerDNS uses the dns.Repository to reply to the dns.Msg `m` with the answer
// in store.Record `r`
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

// StoreHealth uses the health.Repository to generate a health.StoreReport
func (s *LoggedService) StoreHealth() *health.StoreReport {
	s.logger.Log(event.New().
		Level(event.Level_debug).
		Prefix("service").
		Sub("health").
		Message("StoreHealth request").
		Build())

	r := s.svc.StoreHealth()
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

// DNSHealth uses the health.Repository to generate a health.DNSReport
func (s *LoggedService) DNSHealth() *health.DNSReport {
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

// HTTPHealth uses the health.Repository to generate a health.HTTPReport
func (s *LoggedService) HTTPHealth() *health.HTTPReport {
	s.logger.Log(event.New().
		Level(event.Level_debug).
		Prefix("service").
		Sub("health").
		Message("HTTPHealth request").
		Build())

	r := s.svc.HTTPHealth()
	go func() {
		time.Sleep(5 * time.Millisecond)

		s.logger.Log(event.New().
			Level(event.Level_debug).
			Prefix("service").
			Sub("health").
			Message("HTTPHealth response").
			Metadata(event.Field{
				"output": r,
			}).
			Build())
	}()
	return r
}

// Health uses the health.Repository to generate a health.Report
func (s *LoggedService) Health() *health.Report {
	s.logger.Log(event.New().
		Level(event.Level_debug).
		Prefix("service").
		Sub("health").
		Message("MergeHealth request").
		Build())

	r := s.svc.Health()
	go func() {
		time.Sleep(5 * time.Millisecond)

		s.logger.Log(event.New().
			Level(event.Level_debug).
			Prefix("service").
			Sub("health").
			Message("MergeHealth response").
			Metadata(event.Field{
				"output": r,
			}).
			Build())
	}()
	return r
}
