package postgres

import "github.com/zalgonoise/x/gacha/internal/errs"

// Sentinel error kinds and entities used by the PostgreSQL repository.
const (
	domain = "gacha/postgres"

	ErrInvalid = errs.Kind("invalid")

	ErrTableName   = errs.Entity("table name")
	ErrChannelName = errs.Entity("channel name")
)

// Sentinel errors returned by the PostgreSQL repository.
var (
	// ErrInvalidTable is returned for an empty table name, or one with more than a schema and a table part.
	ErrInvalidTable = errs.New(domain, ErrInvalid, ErrTableName)

	// ErrInvalidChannel is returned for an empty notification channel name.
	ErrInvalidChannel = errs.New(domain, ErrInvalid, ErrChannelName)
)
