// internal/repositories/postgres/project_user_postgres.go
package postgres

import (
	"context"

	"authService/internal/domain"
	"authService/internal/repositories"

	"github.com/jackc/pgx/v5/pgxpool"
)

type projectUserRepo struct {
	db *pgxpool.Pool
}

func NewProjectUserRepository(db *pgxpool.Pool) repositories.ProjectUserRepository {
	return &projectUserRepo{db: db}
}

func (r *projectUserRepo) ExistsByEmail(
	ctx context.Context,
	projectID, email string,
) (bool, error) {

	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM "ProjectUser"
			WHERE "projectId" = $1 AND "email" = $2
		)
	`, projectID, email).Scan(&exists)

	return exists, err
}

func (r *projectUserRepo) Create(
	ctx context.Context,
	user *domain.ProjectUser,
) error {

	_, err := r.db.Exec(ctx, `
		INSERT INTO "ProjectUser" (
			id,
			"projectId",
			"providerId",
			"name",
			"email",
			"password",
			"role",
			"tokenVersion",
			"isVerified",
			"createdAt"
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, NOW()
		)
	`,
		user.ID,
		user.ProjectID,
		user.ProviderID,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.TokenVersion,
		user.IsVerified,
	)

	return err
}
