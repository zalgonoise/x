package random

import (
	"context"
	"errors"

	"github.com/zalgonoise/x/randoom/config"
	"github.com/zalgonoise/x/randoom/items"

	"github.com/google/uuid"
)

var ErrUnimplemented = errors.New("unimplemented")

type Repository interface {
	GetRandomItem(ctx context.Context) (*items.Item, error)
	CreatePlaylist(ctx context.Context, label string) (uuid.UUID, []items.Item, error)
	DeletePlaylist(ctx context.Context, id uuid.UUID) error
}

type Service struct {
	repo Repository

	cfg *config.Config
}

func NewService(repo Repository, cfg *config.Config) *Service {
	return &Service{repo: repo, cfg: cfg}
}

func (s *Service) GetRandom() (*items.Item, error) {
	return nil, ErrUnimplemented
}

func (s *Service) NewPlaylist() (uuid.UUID, []items.Item, error) {
	return uuid.UUID{}, nil, ErrUnimplemented
}

func (s *Service) ClosePlaylist(uuid.UUID) error {
	return ErrUnimplemented
}

// TODO: add server start logic
