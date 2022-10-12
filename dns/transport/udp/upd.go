package udp

// Server interface allows launching a UDP server to serve as a DNS server
type Server interface {
	Start() error
	Stop() error
}
