package config

import "os"

type StoreConfig struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}

func StorePath(p string) ConfigOption {
	_, err := os.Stat(p)
	if err != nil {
		return nil
	}
	return &storePath{
		p: p,
	}
}
func StoreType(t string) ConfigOption {
	switch t {
	case "memmap", "memstore":
		return &storeType{
			t: "memmap",
		}
	case "yamlfile", "yaml":
		return &storeType{
			t: "yamlfile",
		}
	case "jsonfile", "json":
		return &storeType{
			t: "jsonfile",
		}
	default:
		return &storeType{
			t: "memmap",
		}
	}
}

type storePath struct {
	p string
}
type storeType struct {
	t string
}

func (l *storePath) Apply(c *Config) {
	c.Store.Path = l.p
}

func (l *storeType) Apply(c *Config) {
	c.Store.Type = l.t
}
