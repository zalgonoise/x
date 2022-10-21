package factory

import (
	"github.com/zalgonoise/x/dns/health"
	"github.com/zalgonoise/x/dns/health/simplehealth"
)

func HealthRepository(rtype string) health.Repository {
	switch rtype {
	case "simple", "simplehealth":
		return simplehealth.New()
	default:
		return simplehealth.New()
	}
}
