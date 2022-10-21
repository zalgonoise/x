package dns

import (
	"errors"

	dnsr "github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/store"
)

var (
	ErrUnimplemented error = errors.New("unimplemented DNS server")
)

type unimplemented struct{}

// Answer implements the dns.Repository interface
func (u unimplemented) Answer(*store.Record, *dnsr.Msg) {}

// Fallback implements the dns.Repository interface
func (u unimplemented) Fallback(*store.Record, *dnsr.Msg) {}

// Unimplemented returns an unimplemented (and invalid) dns.Repository
func Unimplemented() unimplemented {
	return unimplemented{}
}
