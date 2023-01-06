package service

import (
	"context"

	"github.com/zalgonoise/x/secr/secret"
)

func (s service) CreateSecret(ctx context.Context, username string, key string, value []byte) error
func (s service) GetSecret(ctx context.Context, username string, key string) (*secret.Secret, error)
func (s service) ListSecrets(ctx context.Context, username string) ([]*secret.Secret, error)
func (s service) DeleteSecret(ctx context.Context, username string, key string) error
