package core

import (
	"context"
	"fmt"

	dnsr "github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/dns"
	"github.com/zalgonoise/x/dns/store"
	// "github.com/zalgonoise/zlog/log/event"
)

const (
	addr  string = ":53"
	proto string = "udp"
)

type DNSCore struct {
	addr   string
	prefix string
	store  store.ReadRepository
	ctl    chan struct{}
	err    error
}

func New(d *dns.DNS) *DNSCore {
	if d == nil {
		d = dns.New().Build()
	}
	return &DNSCore{
		addr:   d.Addr,
		prefix: d.Prefix,
		store:  store.Unimplemented(),
	}
}

func (d *DNSCore) answerFor(rtype dns.RecordType, question dnsr.Question, m *dnsr.Msg) {
	ctx := context.Background()

	answer, err := d.store.GetByDomain(
		ctx,
		store.New().Type(rtype.String()).Name(question.Name).Build(),
	)
	if err != nil {
		d.err = err
		return
	}
	if answer.Addr != "" {
		response, err := dnsr.NewRR(
			fmt.Sprintf("%s %s %s", question.Name, rtype.String(), answer.Addr),
		)
		if err != nil {
			d.err = err
			return
		}
		m.Answer = append(m.Answer, response)
	}
}

func (d *DNSCore) parseQuery(m *dnsr.Msg) {
	for _, question := range m.Question {
		switch question.Qtype {
		case dnsr.TypeA:
			d.answerFor(dns.TypeA, question, m)
		case dnsr.TypeAAAA:
			d.answerFor(dns.TypeAAAA, question, m)
		case dnsr.TypeCNAME:
			d.answerFor(dns.TypeCNAME, question, m)
		}
	}
}
func (d *DNSCore) HandleRequest(w dnsr.ResponseWriter, r *dnsr.Msg) {
	m := new(dnsr.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dnsr.OpcodeQuery:
		d.parseQuery(m)
		if d.err != nil {
			return
		}
	}

	err := w.WriteMsg(m)
	if err != nil {
		d.err = err
	}
}

func (d *DNSCore) Start() error {
	dnsr.HandleFunc(d.prefix, d.HandleRequest)
	var server = &dnsr.Server{
		Addr: addr,
		Net:  proto,
	}
	d.ctl = make(chan struct{})
	go func() {
		for range d.ctl {
			err := server.Shutdown()
			if err != nil {
				d.err = err
			}
		}
	}()

	return server.ListenAndServe()
}

func (d *DNSCore) Stop() error {
	d.ctl <- struct{}{}

	return d.err
}

func (d *DNSCore) Reload() error {
	err := d.Stop()
	if err != nil {
		return err
	}
	return d.Start()
}

func (d *DNSCore) Store(storeR store.ReadRepository) {
	if storeR == nil && d.store == nil {
		d.store = store.Unimplemented()
		return
	}
	d.store = storeR
}
