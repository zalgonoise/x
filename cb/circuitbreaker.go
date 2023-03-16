package cb

import (
	"errors"
	"sync"
	"time"
)

type err string

func (e err) Error() string {
	return (string)(e)
}

const (
	minFailures    = 2
	minInFlight    = 3
	defaultTimeout = 100 * time.Millisecond

	ErrBreakerIsOpen err = "circuit breaker is open"
)

// state describes the potential states in a CircuitBreaker
type state uint8

const (
	closed   state = iota // closed indicates that the CircuitBreaker allows passing through execution requests
	halfOpen              // halfOpen indicates that the CircuitBreaker is testing sending new execution requests
	open                  // open indicates that the CircuitBreaker has tripped and is currently holding incoming execution requests
)

// ExecFn is a simple type describing an execution request that may fail
type ExecFn func() error

// CircuitBreaker is a data structure that halts the execution of recurrent function if they start raising errors too
// frequently. When this limit is tripped, the CircuitBreaker enters its open-state and waits to retry the requests again
type CircuitBreaker struct {
	state       state
	maxFailures int
	timeout     time.Duration
	errs        []error
	queue       []ExecFn
	success     chan struct{}
	done        chan struct{}
	flusher     func(error)
	mu          sync.Mutex
}

// Exec issues an input execution request ExecFn `fn`, and returns an error if raised.
//
// In case the ExecFn fails, it is queued so it can be retried later
func (c *CircuitBreaker) Exec(fn ExecFn) error {
	if c.state == open {
		c.queue = append(c.queue, fn)
		return ErrBreakerIsOpen
	}

	if e := fn(); e != nil {
		c.onError(fn, e)
		return e
	}

	c.flush()
	c.success <- struct{}{}
	return nil
}

// Close implements the io.Closer interface. It gracefully shuts down the CircuitBreaker
func (c *CircuitBreaker) Close() error {
	c.done <- struct{}{}
	return nil
}

// run is the runtime goroutine for a CircuitBreaker, to keep consuming queue items if existing, if in a non-open state
func (c *CircuitBreaker) run() {
	for {
		select {
		case <-c.done:
			return
		default:
			if len(c.queue) > 0 && c.state < open {
				c.drain()
			}
		}
	}
}

// drain consumes all the present execution functions ExecFn in the CircuitBreaker
func (c *CircuitBreaker) drain() {
	for i := 0; i < len(c.queue); i++ {
		fn := c.queue[0]
		c.queue = c.queue[1:]
		// the error is already handled within c.Exec()
		_ = c.Exec(fn)
	}
}

// flush will reset the CircuitBreaker('s errors), by optionally flushing them with a
// configured flusher `func(error)`, if present
func (c *CircuitBreaker) flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.state = closed
	if len(c.errs) == 0 {
		return
	}
	if c.flusher != nil {
		c.flusher(errors.Join(c.errs...))
	}
	c.errs = c.errs[:0]
}

// onError appends the ExecFn `fn` to the end of the queue, and appends the error to the list
// of errors.
//
// it also handles the state-gate for the CircuitBreaker, whenever the state is not halfOpen or
// the number of failed executions has exceeded the limit
func (c *CircuitBreaker) onError(fn ExecFn, e error) {
	c.mu.Lock()
	c.queue = append(c.queue, fn)
	c.errs = append(c.errs, e)
	if c.state != halfOpen || len(c.errs) > c.maxFailures {
		c.state = open
		defer c.resetAfterTimeout()
	}
	c.mu.Unlock()
}

// resetAfterTimeout resets the CircuitBreaker after the set timeout duration, if the executions start
// working again within the set limits
func (c *CircuitBreaker) resetAfterTimeout() {
	<-time.After(c.timeout)

	c.state = halfOpen
	if len(c.queue) > 0 {
		c.drain()
	}

	c.flush()
}

// NewBufferedBreaker creates a CircuitBreaker with a certain buffer for accepting a maximum number
// of in-flight functions; as well as a maximum number of failures and a timeout duration
func NewBufferedBreaker(maxFailures, maxInFlight int, timeout time.Duration) *CircuitBreaker {
	if maxFailures <= minFailures {
		maxFailures = minFailures
	}
	if maxInFlight <= minInFlight {
		maxInFlight = minInFlight
	}
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	if maxInFlight <= maxFailures {
		// prevent blocking on error by having room for buffering ExecFn
		maxInFlight = maxFailures
	}

	c := &CircuitBreaker{
		maxFailures: maxFailures,
		timeout:     timeout,
		errs:        make([]error, 0, maxFailures),
		queue:       make([]ExecFn, 0, maxInFlight),
		success:     make(chan struct{}),
		done:        make(chan struct{}),
	}

	go c.run()
	return c
}

// NewCircuitBreaker creates a CircuitBreaker with a certain buffer for accepting a maximum number
// of failures and a timeout duration
func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	if maxFailures <= minFailures {
		maxFailures = minFailures
	}
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	c := &CircuitBreaker{
		maxFailures: maxFailures,
		timeout:     timeout,
		errs:        make([]error, 0, maxFailures),
		queue:       make([]ExecFn, 0, minFailures),
		success:     make(chan struct{}),
		done:        make(chan struct{}),
	}

	go c.run()
	return c
}

// WithFlusher sets a function to consume the joined errors each time a CircuitBreaker is drained
func (c *CircuitBreaker) WithFlusher(fn func(error)) *CircuitBreaker {
	c.flusher = fn
	return c
}

// Results returns a channel that reports whenever an execution is successful
func (c *CircuitBreaker) Results() <-chan struct{} {
	return c.success
}
