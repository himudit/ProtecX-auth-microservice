package domain

import "time"

type ProjectRole string

const (
	Owner  ProjectRole = "OWNER"
	Admin  ProjectRole = "ADMIN"
	Member ProjectRole = "MEMBER"
)

type ProjectUser struct {
	ID         string
	ProjectID  string
	ProviderID string

	Name         string
	Email        string
	PasswordHash string
	Role         ProjectRole
	IsVerified   bool
	TokenVersion int

	CreatedAt   time.Time
	LastLoginAt *time.Time
}
