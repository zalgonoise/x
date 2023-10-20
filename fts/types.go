package fts

import "database/sql"

type Number interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 |
		float32 | float64
}

type Char interface {
	string | []byte
}

type SQLNullable interface {
	sql.NullBool |
		sql.NullInt16 | sql.NullInt32 | sql.NullInt64 |
		sql.NullString
}

type SQLType interface {
	Number | Char
}
