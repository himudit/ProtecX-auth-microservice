package repositories

import (
	"context"

	"authService/internal/domain"
)

type ProjectJwtKeyRepository interface {
	GetActiveKeyByProjectID(ctx context.Context, projectID string) (*domain.ProjectJwtKey, error)
}
