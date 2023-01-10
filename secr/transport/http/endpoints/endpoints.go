package endpoints

import (
	"github.com/zalgonoise/x/secr/service"
	"github.com/zalgonoise/x/secr/transport/http"
)

type endpoints struct {
	s service.Service
}

func NewAPI(s service.Service) http.API {
	return endpoints{s}
}
