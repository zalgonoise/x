package endpoints

import (
	"github.com/zalgonoise/x/secr/service"
	"github.com/zalgonoise/x/secr/transport/http"
)

type endpoints struct {
	s   service.Service
	enc EncodeDecoder
}

func NewAPI(s service.Service, encoderType string) http.API {
	return endpoints{
		s:   s,
		enc: NewEncoder(encoderType),
	}
}
