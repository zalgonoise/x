package config

import "os"

type StoreConfig struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}

// StorePath creates a ConfigOption setting the Config's store path to string `p`
//
// It tries to call os.Stat() on string `p` to evaluate if the file exists.
// If it doesn't, it attempts to create it. If that fails too the returned ConfigOption
// is `nil`; otherwise it will return a ConfigOption to update the store path.
func StorePath(p string) ConfigOption {
	_, err := os.Stat(p)
	if err != nil {
		f, err := os.Create(p)
		if err != nil {
			return nil
		}
		err = f.Sync()
		if err != nil {
			return nil
		}
	}
	return &storePath{
		p: p,
	}
}

// StoreType creates a ConfigOption setting the Config's store type to string `t`
//
// It defaults to `memmap`
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

// Apply implements the ConfigOption interface
func (l *storePath) Apply(c *Config) {
	c.Store.Path = l.p
}

// Apply implements the ConfigOption interface
func (l *storeType) Apply(c *Config) {
	c.Store.Type = l.t
}
