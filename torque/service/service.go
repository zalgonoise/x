package service

import (
	"context"
	"log/slog"

	"github.com/zalgonoise/x/torque/vehicles"
)

type Repository interface {
	InsertVehicle(ctx context.Context, v vehicles.Vehicle) error
	BulkInsertVehicles(ctx context.Context, v []vehicles.Vehicle) error
	InsertHandling(ctx context.Context, v vehicles.Handling) error
	BulkInsertHandling(ctx context.Context, v []vehicles.Handling) error
}

type Service struct {
	repo Repository

	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}
