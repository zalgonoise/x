package miekgdns

import "github.com/miekg/dns"

// Start launches the DNS server, returning an error
func (u *udps) Start() error {
	if u.Running() {
		return ErrAlreadyRunning
	}
	dns.HandleFunc(u.conf.Prefix, u.handleRequest)
	u.srv = &dns.Server{
		Addr: u.conf.Addr,
		Net:  u.conf.Proto,
	}

	return u.srv.ListenAndServe()
}

// Stop gracefully stops the DNS server, returning an error
func (u *udps) Stop() error {
	if !u.Running() {
		return ErrNotRunning
	}
	return u.srv.Shutdown()
}

// Running returns a boolean on whether the UDP server is running or not
func (u *udps) Running() bool {
	if u.srv != nil {
		if addr := u.srv.Listener.Addr(); addr != nil {
			return true
		}
	}
	return false
}
