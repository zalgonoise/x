package service

import (
	"fmt"

	"github.com/zalgonoise/x/errors"
)

type Transactioner interface {
	Rollback(error) error
	Add(RollbackFn)
}

type RollbackFn func() error

type transactioner struct {
	r []RollbackFn
}

func newTx() Transactioner {
	return &transactioner{}
}

func (tx *transactioner) Rollback(input error) error {
	var errs = make([]error, 0, len(tx.r)+1)
	errs[0] = input

	for _, rb := range tx.r {
		if err := rb(); err != nil {
			errs = append(errs, fmt.Errorf("rollback error: %w", err))
		}
	}
	switch len(errs) {

	case 1:
		return errs[0]
	default:
		return errors.Join(errs...)
	}
}

func (tx *transactioner) Add(r RollbackFn) {
	tx.r = append(tx.r, r)
}
