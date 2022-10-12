package miekgdns

import (
	dnsr "github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/dns"
	"github.com/zalgonoise/x/dns/service"
	"github.com/zalgonoise/x/dns/store"
)

// miekgdns implements the udp.Server interface
type udps struct {
	service service.Service
	conf    *dns.DNS
	ctl     chan struct{}
	err     error
}

func New(conf *dns.DNS, s service.Service) *udps {
	if conf == nil {
		conf = dns.New().Build()
	}
	return &udps{
		conf:    conf,
		service: s,
	}
}

func (u *udps) Start() error {
	dnsr.HandleFunc(u.conf.Prefix, u.HandleRequest)
	var server = &dnsr.Server{
		Addr: u.conf.Addr,
		Net:  u.conf.Proto,
	}
	// launch shutdown controller
	u.ctl = make(chan struct{})
	go func() {
		for range u.ctl {
			err := server.Shutdown()
			if err != nil {
				u.err = err
			}
		}
	}()

	return server.ListenAndServe()
}

func (u *udps) Stop() error {
	u.ctl <- struct{}{}

	return u.err
}

func (u *udps) Answer(r *store.Record, m *dnsr.Msg) {
	u.service.AnswerDNS(r, m)
}

func (u *udps) HandleRequest(w dnsr.ResponseWriter, r *dnsr.Msg) {
	m := new(dnsr.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dnsr.OpcodeQuery:
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

func (u *udps) parseQuery(m *dnsr.Msg) {
	for _, question := range m.Question {
		r := store.New().Name(question.Name)
		switch question.Qtype {
		case dnsr.TypeA:
			u.Answer(
				r.Type(dns.TypeA.String()).Build(),
				m,
			)
		case dnsr.TypeAAAA:
			u.Answer(
				r.Type(dns.TypeAAAA.String()).Build(),
				m,
			)
		case dnsr.TypeCNAME:
			u.Answer(
				r.Type(dns.TypeCNAME.String()).Build(),
				m,
			)
		}
	}
}
