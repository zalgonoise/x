package flags

import (
	"flag"
	"os"

	"github.com/zalgonoise/x/secr/cmd/config"
)

func ParseFlags() *config.Config {
	var conf = &config.Default

	boltDBPath := flag.String("bolt-path", conf.BoltDBPath, "path to the Bolt database file")
	sqliteDBPath := flag.String("sqlite-path", conf.SQLiteDBPath, "path to the SQLite database file")
	signingKeyPath := flag.String("jwt-key", conf.SigningKeyPath, "path to the JWT signing key file")
	logfilePath := flag.String("logfile-path", conf.LogFilePath, "path to the logfile stored in the service")
	tracefilePath := flag.String("tracefile-path", conf.TraceFilePath, "path to the tracefile stored in the service")

	flag.Parse()
	osFlags := ParseOSEnv()

	conf.Apply(
		config.BoltDB(*boltDBPath),
		config.SQLiteDB(*sqliteDBPath),
		config.JWTKey(*signingKeyPath),
		config.Logfile(*logfilePath),
		config.Tracefile(*tracefilePath),
	)

	return config.Merge(conf, osFlags)
}

func ParseOSEnv() *config.Config {
	return &config.Config{
		BoltDBPath:     os.Getenv("SECR_BOLT_PATH"),
		SQLiteDBPath:   os.Getenv("SECR_SQLITE_PATH"),
		SigningKeyPath: os.Getenv("SECR_JWT_KEY_PATH"),
		LogFilePath:    os.Getenv("SECR_LOGFILE_PATH"),
		TraceFilePath:  os.Getenv("SECR_TRACEFILE_PATH"),
	}
}
