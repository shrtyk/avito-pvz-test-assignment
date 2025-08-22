package auth

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	UserRoleEmployee  UserRole = "employee"
	UserRoleModerator UserRole = "moderator"
)

type User struct {
	Id           uuid.UUID
	PasswordHash []byte
	Email        string
	Role         UserRole
	CreatedAt    time.Time
}

type RegisterUserParams struct {
	Email         string
	PlainPassword string
	Role          UserRole
}
