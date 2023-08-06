package service

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"
)

var ErrOverThreshold = errors.New("value is over threshold")

// Service describes the behavior that a general service (as a handler) would have
//
// It is declared on the same file as the implementation simply to make it easer for
// the observability wrappers
type Service interface {
	Handle(ctx context.Context, value int) error
}

type Handler struct {
	Threshold int

	rand *rand.Rand
}

func (h Handler) Handle(_ context.Context, value int) error {
	time.Sleep(time.Duration(h.rand.Intn(100)) * time.Millisecond)

	if value >= h.Threshold {
		return ErrOverThreshold
	}

	return nil
}

func NewHandler(threshold int) Handler {
	return Handler{
		Threshold: threshold,
		rand: rand.New(rand.NewSource(
			int64(float64(time.Now().Unix()) / math.Pi),
		)),
	}
}
