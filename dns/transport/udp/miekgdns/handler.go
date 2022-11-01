package miekgdns

import (
	"github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/store"
)

func (u *udps) answer(r *store.Record, m *dns.Msg) {
	name := r.Name
	if r.Name[len(r.Name)-1] == u.conf.Prefix[0] {
		name = r.Name[:len(r.Name)-1]
	}

	u.ans.AnswerDNS(
		store.New().Name(name).Type(r.Type).Build(),
		m,
	)
}

func (u *udps) handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		u.parseQuery(m)
		if u.err != nil {
			return
		}
	}

	err := w.WriteMsg(m)
	if err != nil {
		u.err = err
	}
}

func (u *udps) parseQuery(m *dns.Msg) {
	for _, question := range m.Question {
		r := store.New().Name(question.Name)
		switch question.Qtype {
		case dns.TypeA:
			u.answer(
				r.Type(store.TypeA.String()).Build(),
				m,
			)
		case dns.TypeAAAA:
			u.answer(
				r.Type(store.TypeAAAA.String()).Build(),
				m,
			)
		case dns.TypeCNAME:
			u.answer(
				r.Type(store.TypeCNAME.String()).Build(),
				m,
			)
		case dns.TypeANY:
			u.answer(
				r.Build(),
				m,
			)
		}
	}
}
