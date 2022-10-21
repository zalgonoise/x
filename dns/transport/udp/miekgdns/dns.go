package miekgdns

import (
	"errors"

	dnsr "github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/service"
	"github.com/zalgonoise/x/dns/store"
	"github.com/zalgonoise/x/dns/transport/udp"
)

var (
	ErrAlreadyRunning error = errors.New("DNS server is already running")
	ErrNotRunning     error = errors.New("DNS server is not running, yet")
)

// miekgdns implements the udp.Server interface
type udps struct {
	ans  service.Answering
	conf *udp.DNS
	srv  *dnsr.Server
	err  error
}

func NewServer(conf *udp.DNS, s service.Service) udp.Server {
	if conf == nil {
		conf = udp.NewDNS().Build()
	}
	return &udps{
		conf: conf,
		ans:  s,
	}
}

func (u *udps) Start() error {
	if u.Running() {
		return ErrAlreadyRunning
	}
	dnsr.HandleFunc(u.conf.Prefix, u.handleRequest)
	u.srv = &dnsr.Server{
		Addr: u.conf.Addr,
		Net:  u.conf.Proto,
	}

	return u.srv.ListenAndServe()
}

func (u *udps) Stop() error {
	if !u.Running() {
		return ErrNotRunning
	}
	return u.srv.Shutdown()
}

func (u *udps) Running() bool {
	if u.srv != nil {
		if addr := u.srv.Listener.Addr(); addr != nil {
			return true
		}
	}
	return false
}

func (u *udps) answer(r *store.Record, m *dnsr.Msg) {
	name := r.Name
	if r.Name[len(r.Name)-1] == u.conf.Prefix[0] {
		name = r.Name[:len(r.Name)-1]
	}

	u.ans.AnswerDNS(
		store.New().Name(name).Type(r.Type).Build(),
		m,
	)
}

func (u *udps) handleRequest(w dnsr.ResponseWriter, r *dnsr.Msg) {
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
			u.answer(
				r.Type(store.TypeA.String()).Build(),
				m,
			)
		case dnsr.TypeAAAA:
			u.answer(
				r.Type(store.TypeAAAA.String()).Build(),
				m,
			)
		case dnsr.TypeCNAME:
			u.answer(
				r.Type(store.TypeCNAME.String()).Build(),
				m,
			)
		}
	}
}
