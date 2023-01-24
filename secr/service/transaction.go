package service

import "fmt"

type Transactioner interface {
	Rollback(error) error
	Add(RollbackFn)
}

type RollbackFn func() error

type transactioner struct {
	r   []RollbackFn
	err error
}

func newTx() Transactioner {
	return &transactioner{}
}

func (tx *transactioner) Rollback(input error) error {
	for _, rb := range tx.r {
		err := rb()
		if err != nil {
			if tx.err == nil {
				tx.err = err
				continue
			}
			tx.err = fmt.Errorf("%w -- %v", tx.err, err)
		}
	}
	if tx.err == nil {
		return input
	}
	return fmt.Errorf("%w -- rollback error: %v", input, tx.err)
}

func (tx *transactioner) Add(r RollbackFn) {
	tx.r = append(tx.r, r)
}
