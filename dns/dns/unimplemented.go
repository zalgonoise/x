package dns

import (
	"errors"

	"github.com/zalgonoise/x/dns/store"
)

var (
	ErrUnimplemented error = errors.New("unimplemented DNS server")
)

type unimplemented struct{}

// func (u unimplemented) HandleRequest(dns.ResponseWriter, *dns.Msg) {}

func (u unimplemented) Start() error {
	return ErrUnimplemented
}

func (u unimplemented) Stop() error {
	return ErrUnimplemented
}

func (u unimplemented) Reload() error {
	return ErrUnimplemented
}

func (u unimplemented) Link() chan *store.Record {
	return nil
}

func Unimplemented() unimplemented {
	return unimplemented{}
}
