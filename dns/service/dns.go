package service

import (
	"context"

	dnsr "github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/store"
)

// AnswerDNS uses the dns.Repository to reply to the dns.Msg `m` with the answer
// in store.Record `r`
func (s *service) AnswerDNS(r *store.Record, m *dnsr.Msg) {
	var (
		ctx = context.Background()
	)
	switch r.Type {
	case "", "ANY":
		answers, err := s.store.FilterByDomain(ctx, r.Name)
		if err != nil || len(answers) == 0 {
			r.Type = "ANY"
			s.dns.Fallback(r, m)
			return
		}

		for _, ans := range answers {
			s.dns.Answer(ans, m)
		}
	default:
		answer, err := s.store.FilterByTypeAndDomain(ctx, r.Type, r.Name)
		if err != nil || answer.Addr == "" {
			s.dns.Fallback(r, m)
			return
		}

		s.dns.Answer(answer, m)
	}
}
