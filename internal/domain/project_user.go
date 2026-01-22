package domain

import "time"

type ProjectUser struct {
	ID         string
	ProjectID  string
	ProviderID string

	Name         string
	Email        string
	PasswordHash string
	Role         string
	IsVerified   bool
	TokenVersion int

	CreatedAt   time.Time
	LastLoginAt *time.Time
}
