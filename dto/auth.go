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

type LoginResponse struct {
	AccessToken string `json:"access_token" binding:"required"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
