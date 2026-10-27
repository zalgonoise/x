package mysql

import "github.com/zalgonoise/x/gacha/internal/errs"

// Sentinel error kinds and entities used by the MySQL repository.
const (
	domain = "gacha/mysql"

	ErrInvalid = errs.Kind("invalid")

	ErrTableName   = errs.Entity("table name")
	ErrColumnValue = errs.Entity("column value")
)

// Sentinel errors returned by the MySQL repository.
var (
	// ErrInvalidTable is returned for an empty table name, or one with more than a database and a table part.
	ErrInvalidTable = errs.New(domain, ErrInvalid, ErrTableName)

	// ErrInvalidValue is returned when a column holds a value of an unexpected type, such as a table that was not
	// created by Migrate or Schema.
	ErrInvalidValue = errs.New(domain, ErrInvalid, ErrColumnValue)
)
