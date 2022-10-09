package uimpdns

import (
	"errors"

	"github.com/miekg/dns"
)

var (
	ErrUnimplementedDNS error = errors.New("unimplemented DNS server")
)

type UnimplementedDNS struct{}

func (u UnimplementedDNS) ParseQuery(*dns.Msg) {}

func (u UnimplementedDNS) HandleRequest(dns.ResponseWriter, *dns.Msg) {}

func (u UnimplementedDNS) Start() error {
	return ErrUnimplementedDNS
}

func (u UnimplementedDNS) Stop() error {
	return ErrUnimplementedDNS
}

func (u UnimplementedDNS) Reload() error {
	return ErrUnimplementedDNS
}

func New() UnimplementedDNS {
	return UnimplementedDNS{}
}
