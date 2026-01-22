package postgres

import (
	"context"

	"authService/internal/domain"
	"authService/internal/repositories"

	"github.com/jackc/pgx/v5/pgxpool"
)

type projectJwtKeyRepo struct {
	db *pgxpool.Pool
}

func NewProjectJwtKeyRepository(db *pgxpool.Pool) repositories.ProjectJwtKeyRepository {
	return &projectJwtKeyRepo{db: db}
}

func (r *projectJwtKeyRepo) GetActiveKeyByProjectID(ctx context.Context, projectID string) (*domain.ProjectJwtKey, error) {
	var key domain.ProjectJwtKey
	err := r.db.QueryRow(ctx, `
		SELECT 
			id, project_id, kid, public_key, private_key_encrypted, algorithm, is_active, created_at
		FROM project_jwt_keys
		WHERE project_id = $1 AND is_active = true
		LIMIT 1
	`, projectID).Scan(
		&key.ID,
		&key.ProjectID,
		&key.Kid,
		&key.PublicKey,
		&key.PrivateKeyEncrypted,
		&key.Algorithm,
		&key.IsActive,
		&key.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &key, nil
}
