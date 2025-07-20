package dto

import "github.com/vnFuhung2903/vcs-sms/entities"

type RegisterRequest struct {
	Username string            `json:"username" binding:"required"`
	Password string            `json:"password" binding:"required"`
	Email    string            `json:"email" binding:"required,email"`
	Role     entities.UserRole `json:"role" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdatePasswordRequest struct {
	Password string `json:"password"`
}

type UpdateRoleRequest struct {
	Role entities.UserRole `json:"role" binding:"required"`
}

type UpdateScopeRequest struct {
	IsAdded bool     `json:"is_added"`
	Scopes  []string `json:"scopes" binding:"required"`
}
