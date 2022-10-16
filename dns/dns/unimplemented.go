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

func (u unimplemented) Answer(*store.Record, *dnsr.Msg) {}

func (u unimplemented) Fallback(*store.Record, *dnsr.Msg) {}

func Unimplemented() unimplemented {
	return unimplemented{}
}
