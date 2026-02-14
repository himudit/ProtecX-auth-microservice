// internal/repositories/postgres/project_user_postgres.go
package postgres

import (
	"context"

	"authService/internal/domain"
	"authService/internal/repositories"

	"github.com/jackc/pgx/v5"
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

func (r *projectUserRepo) GetUserByEmail(
	ctx context.Context,
	projectID, email string,
) (*domain.ProjectUser, error) {

	user := &domain.ProjectUser{}
	err := r.db.QueryRow(ctx, `
        SELECT 
		"id", 
		"projectId",
		"providerId",
		"name",
		"email",
		"password",
		"role",
		"tokenVersion",
		"isVerified",
		"createdAt",
		"lastLoginAt"
        FROM "ProjectUser"
        WHERE "projectId" = $1 AND "email" = $2
    `, projectID, email).Scan(
		&user.ID,
		&user.ProjectID,
		&user.ProviderID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.TokenVersion,
		&user.IsVerified,
		&user.CreatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows { // if using pgx
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}
func (r *projectUserRepo) GetUserByID(
	ctx context.Context,
	projectID, userID string,
) (*domain.ProjectUser, error) {

	user := &domain.ProjectUser{}
	err := r.db.QueryRow(ctx, `
        SELECT 
		"id", 
		"projectId",
		"providerId",
		"name",
		"email",
		"password",
		"role",
		"tokenVersion",
		"isVerified",
		"createdAt",
		"lastLoginAt"
        FROM "ProjectUser"
        WHERE "projectId" = $1 AND "id" = $2
    `, projectID, userID).Scan(
		&user.ID,
		&user.ProjectID,
		&user.ProviderID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.TokenVersion,
		&user.IsVerified,
		&user.CreatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (r *projectUserRepo) IncrementTokenVersion(ctx context.Context, projectID, userID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE "ProjectUser"
		SET "tokenVersion" = "tokenVersion" + 1
		WHERE "projectId" = $1 AND "id" = $2
	`, projectID, userID)
	return err
}
