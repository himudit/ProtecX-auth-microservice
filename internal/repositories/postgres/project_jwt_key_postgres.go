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
			id, "projectId", kid, "publicKey", "privateKeyEncrypted", algorithm, "isActive", "createdAt"
		FROM "projectJwtKeys"
		WHERE "projectId" = $1 AND "isActive" = true
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
