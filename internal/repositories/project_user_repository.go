package repositories

import (
	"context"

	"authService/internal/domain"
)

type ProjectUserRepository interface {
	ExistsByEmail(ctx context.Context, projectID, email string) (bool, error)
	Create(ctx context.Context, user *domain.ProjectUser) error
}
