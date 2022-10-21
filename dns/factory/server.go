package factory

import (
	"github.com/zalgonoise/x/dns/service"
	"github.com/zalgonoise/x/dns/transport/httpapi"
	"github.com/zalgonoise/x/dns/transport/httpapi/endpoints"
	"github.com/zalgonoise/x/dns/transport/udp"
	"github.com/zalgonoise/x/dns/transport/udp/miekgdns"
)

func UDPServer(stype, address, prefix, proto string, svc service.Service) udp.Server {
	var udps udp.Server

	switch stype {
	case "miekgdns":
		udps = miekgdns.NewServer(
			udp.NewDNS().
				Addr(address).
				Prefix(prefix).
				Proto(proto).
				Build(),
			svc,
		)
	default:
		udps = miekgdns.NewServer(
			udp.NewDNS().
				Addr(address).
				Prefix(prefix).
				Proto(proto).
				Build(),
			svc,
		)
	}

	return udps
}

func Server(dnstype, dnsAddress, dnsPrefix, dnsProto string, httpPort int, svc service.Service) (httpapi.Server, udp.Server) {
	udps := UDPServer(dnstype, dnsAddress, dnsPrefix, dnsProto, svc)
	apis := endpoints.NewAPI(svc, udps)
	https := httpapi.NewServer(apis, httpPort)

	return https, udps
}
