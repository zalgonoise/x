package store

import "errors"

var (
	ErrNoAddr           error = errors.New("no IP address provided")
	ErrNoName           error = errors.New("no domain name provided")
	ErrNoType           error = errors.New("no DNS record type provided")
	ErrNotFound         error = errors.New("entry was not found")
	ErrDoesNotExist     error = errors.New("record does not exist")
	ErrAlreadyExists    error = errors.New("entry already exists")
	ErrZeroBytesWritten error = errors.New("zero bytes written")
	ErrSync             error = errors.New("sync error")
	ErrZeroRecords      error = errors.New("zero records in the store")
)
