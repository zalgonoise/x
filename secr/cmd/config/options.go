package config

type boltPath string

func (p boltPath) Apply(c *Config) {
	c.BoltDBPath = (string)(p)
}

func BoltDB(path string) ConfigOption {
	if path == "" {
		return nil
	}
	return (boltPath)(path)
}

type sqlitePath string

func (p sqlitePath) Apply(c *Config) {
	c.SQLiteDBPath = (string)(p)
}

func SQLiteDB(path string) ConfigOption {
	if path == "" {
		return nil
	}
	return (sqlitePath)(path)
}

type jwtKeyPath string

func (p jwtKeyPath) Apply(c *Config) {
	c.SigningKeyPath = (string)(p)
}

func JWTKey(path string) ConfigOption {
	if path == "" {
		return nil
	}
	return (jwtKeyPath)(path)
}

type httpPort int

func (p httpPort) Apply(c *Config) {
	c.HTTPPort = (int)(p)
}

func Port(port int) ConfigOption {
	if port == 0 {
		return nil
	}
	return (httpPort)(port)
}
