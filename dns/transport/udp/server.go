package udp

import "errors"

var (
	ErrAlreadyRunning error = errors.New("DNS server is already running")
	ErrNotRunning     error = errors.New("DNS server is not running, yet")
)

// Server interface allows launching a UDP server
// to use as a DNS server
type Server interface {
	// Start launches the DNS server, returning an error
	Start() error
	// Stop gracefully stops the DNS server, returning an error
	Stop() error
	// Running returns a boolean on whether the UDP server is running or not
	Running() bool
}
