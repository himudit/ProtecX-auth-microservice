package domain

import "time"

type ProjectUser struct {
	ID           string      `json:"id"`
	ProjectID    string      `json:"-"`
	ProviderID   string      `json:"-"`
	Name         string      `json:"name"`
	Email        string      `json:"email"`
	PasswordHash string      `json:"-"`
	Role         ProjectRole `json:"role"`
	IsVerified   bool        `json:"isVerified"`
	TokenVersion int         `json:"-"`
	CreatedAt    time.Time   `json:"-"`
	LastLoginAt  *time.Time  `json:"lastLoginAt"`
}
