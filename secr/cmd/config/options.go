package config

type boltPath string

// Apply sets the configuration on the input Config `c`
func (p boltPath) Apply(c *Config) {
	c.BoltDBPath = (string)(p)
}

// BoltDB defines the path for the Bolt database file
func BoltDB(path string) Option {
	if path == "" {
		return nil
	}
	return (boltPath)(path)
}

type sqlitePath string

// Apply sets the configuration on the input Config `c`
func (p sqlitePath) Apply(c *Config) {
	c.SQLiteDBPath = (string)(p)
}

// SQLiteDB defines the path for the SQLite database file
func SQLiteDB(path string) Option {
	if path == "" {
		return nil
	}
	return (sqlitePath)(path)
}

type jwtKeyPath string

// Apply sets the configuration on the input Config `c`
func (p jwtKeyPath) Apply(c *Config) {
	c.SigningKeyPath = (string)(p)
}

// JWTKey defines the path for the JWT signing key file
func JWTKey(path string) Option {
	if path == "" {
		return nil
	}
	return (jwtKeyPath)(path)
}

type httpPort int

// Apply sets the configuration on the input Config `c`
func (p httpPort) Apply(c *Config) {
	c.HTTPPort = (int)(p)
}

// Port defines the HTTP port for the server
func Port(port int) Option {
	if port == 0 {
		return nil
	}
	return (httpPort)(port)
}

type logfilePath string

// Apply sets the configuration on the input Config `c`
func (p logfilePath) Apply(c *Config) {
	c.LogFilePath = (string)(p)
}

// Logfile defines the path for the error log file
func Logfile(path string) Option {
	if path == "" {
		return nil
	}
	return (logfilePath)(path)
}

type tracefilePath string

// Apply sets the configuration on the input Config `c`
func (p tracefilePath) Apply(c *Config) {
	c.TraceFilePath = (string)(p)
}

// Tracefile defines the path for the trace file
func Tracefile(path string) Option {
	if path == "" {
		return nil
	}
	return (tracefilePath)(path)
}
