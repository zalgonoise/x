package dns

import (
	"errors"

	"github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/store"
)

var (
	ErrUnimplemented error = errors.New("unimplemented DNS server")
)

type unimplemented struct{}

func (u unimplemented) ParseQuery(*dns.Msg) {}

func (u unimplemented) HandleRequest(dns.ResponseWriter, *dns.Msg) {}

func (u unimplemented) Start() error {
	return ErrUnimplemented
}

func (u unimplemented) Stop() error {
	return ErrUnimplemented
}

func (u unimplemented) Reload() error {
	return ErrUnimplemented
}

func (u unimplemented) Store(store.Repository) {}

func Unimplemented() unimplemented {
	return unimplemented{}
}
