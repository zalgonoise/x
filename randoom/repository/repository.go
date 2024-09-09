package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"github.com/zalgonoise/x/randoom/items"
)

var ErrUnimplemented = errors.New("unimplemented")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Load(ctx context.Context, list items.List) error {
	return ErrUnimplemented
}

// TODO: add a Register method for manual, non-randomized increments

func (r *Repository) GetRandomItem(ctx context.Context) (*items.Item, error) {
	return nil, ErrUnimplemented
}

func (r *Repository) CreatePlaylist(ctx context.Context, label string) (uuid.UUID, []items.Item, error) {
	return uuid.UUID{}, nil, ErrUnimplemented
}

func (r *Repository) DeletePlaylist(ctx context.Context, id uuid.UUID) error {
	return ErrUnimplemented
}
