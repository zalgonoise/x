package service

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"

	"github.com/zalgonoise/x/modupdate/config"
)

var (
	ErrNilRuntime    = errors.New("nil cron runtime is not acceptable")
	ErrNilRepository = errors.New("nil repository is not acceptable")
)

type Repository interface {
	ListTasks(ctx context.Context) ([]*config.Task, error)
	AddTask(ctx context.Context, cfg *config.Task) error
	DeleteTask(ctx context.Context, uri, module, branch string) error
	Close() error
}

type Runtime interface {
	Run(ctx context.Context)
	Err() <-chan error
}

// Service handles requests to list, add and remove tasks as stored in the repository,
// as well as keeping a cron instance running to execute the configured tasks.
//
// Modifying the tasks in a service will cause the cron instance to be shut down and restarted.
type Service struct {
	runtime Runtime
	repo    Repository

	logger *slog.Logger

	cancel context.CancelFunc
	err    atomic.Pointer[error]
}

func NewService(runtime Runtime, repo Repository, logger *slog.Logger) (*Service, error) {
	if runtime == nil {
		return nil, ErrNilRuntime
	}

	if repo == nil {
		return nil, ErrNilRepository
	}

	return &Service{
		runtime: runtime,
		repo:    repo,
		logger:  logger,
	}, nil
}

func (s *Service) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	go func() {
		for {
			select {
			case <-ctx.Done():
			case err := <-s.runtime.Err():
				s.err.Store(&err)
				s.logger.ErrorContext(ctx, "cron runtime error", slog.String("error", err.Error()))
			}
		}
	}()

	s.runtime.Run(ctx)
}

func (s *Service) Err() error {
	err := s.err.Load()
	if err == nil {
		return nil
	}

	return *err
}

func (s *Service) Close() error {
	s.cancel()

	return s.repo.Close()
}
