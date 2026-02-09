package domain

type ProjectRole string

const (
	RoleOwner  ProjectRole = "OWNER"
	RoleAdmin  ProjectRole = "ADMIN"
	RoleMember ProjectRole = "MEMBER"
)
