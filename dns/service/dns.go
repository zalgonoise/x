package service

import (
	"context"

	dnsr "github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/store"
)

func (s *service) AnswerDNS(r *store.Record, m *dnsr.Msg) {
	ctx := context.Background()
	answer, err := s.store.FilterByDomain(ctx, r)
	if err != nil || answer.Addr == "" {
		s.dns.Fallback(r, m)
		return
	}

	s.dns.Answer(answer, m)
}
