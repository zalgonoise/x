package miekgdns

import (
	"github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/transport/udp"
)

// Start launches the DNS server, returning an error
func (u *udps) Start() error {
	if u.on {
		return udp.ErrAlreadyRunning
	}
	dns.HandleFunc(u.conf.Prefix, u.handleRequest)
	u.srv = &dns.Server{
		Addr: u.conf.Addr,
		Net:  u.conf.Proto,
	}
	u.on = true

	return u.srv.ListenAndServe()
}

// Stop gracefully stops the DNS server, returning an error
func (u *udps) Stop() error {
	if !u.on {
		return udp.ErrNotRunning
	}
	u.on = false
	return u.srv.Shutdown()
}

// Running returns a boolean on whether the UDP server is running or not
func (u *udps) Running() bool {
	return u.on
}
