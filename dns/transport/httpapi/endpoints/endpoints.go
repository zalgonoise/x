package endpoints

import (
	"github.com/zalgonoise/x/dns/service"
	"github.com/zalgonoise/x/dns/store/encoder"
	"github.com/zalgonoise/x/dns/transport/httpapi"
	"github.com/zalgonoise/x/dns/transport/udp"
)

type endpoints struct {
	s   service.StoreWithHealth
	UDP udp.Server
	enc encoder.EncodeDecoder
}

func NewAPI(s service.Service, udps udp.Server) httpapi.HTTPAPI {
	return &endpoints{
		s:   s,
		UDP: udps,
		enc: encoder.New("json"),
	}
}
