package factory

import (
	"github.com/zalgonoise/x/dns/store"
	"github.com/zalgonoise/x/dns/store/jsonfile"
	"github.com/zalgonoise/x/dns/store/memmap"
	"github.com/zalgonoise/x/dns/store/yamlfile"
)

func StoreRepository(rtype string, path string) store.Repository {
	var storeRepo store.Repository

	switch rtype {
	case "memmap", "memory", "in-memory":
		storeRepo = memmap.New()
	case "jsonfile", "json":
		storeRepo = jsonfile.New(path)
	case "yamlfile", "yaml":
		storeRepo = yamlfile.New(path)
	default:
		storeRepo = memmap.New()
	}

	return storeRepo
}
