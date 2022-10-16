package dns

import (
	dnsr "github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/store"
)

type Repository interface {
	Answer(*store.Record, *dnsr.Msg)
	Fallback(*store.Record, *dnsr.Msg)
}
