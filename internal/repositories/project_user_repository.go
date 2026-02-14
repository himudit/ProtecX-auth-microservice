package repositories

import (
	"context"

	"authService/internal/domain"
)

type ProjectUserRepository interface {
	ExistsByEmail(ctx context.Context, projectID, email string) (bool, error)
	Create(ctx context.Context, user *domain.ProjectUser) error
	GetUserByEmail(ctx context.Context, projectID, email string) (*domain.ProjectUser, error)
	GetUserByID(ctx context.Context, projectID, userID string) (*domain.ProjectUser, error)
	IncrementTokenVersion(ctx context.Context, projectID, userID string) error
}
