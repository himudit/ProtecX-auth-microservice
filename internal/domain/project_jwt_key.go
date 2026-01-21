package domain

import "time"

type ProjectJwtKey struct {
	ID                  string
	ProjectID           string
	Kid                 string
	PublicKey           string
	PrivateKeyEncrypted string
	Algorithm            string
	IsActive            bool
	CreatedAt            time.Time
}
