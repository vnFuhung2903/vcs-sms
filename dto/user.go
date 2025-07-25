package dto

import "github.com/vnFuhung2903/vcs-sms/entities"

type UpdateRoleRequest struct {
	UserId string            `json:"user_id" binding:"required"`
	Role   entities.UserRole `json:"role" binding:"required"`
}

type UpdateScopeRequest struct {
	UserId  string   `json:"user_id" binding:"required"`
	IsAdded bool     `json:"is_added"`
	Scopes  []string `json:"scopes" binding:"required"`
}

type DeleteRequest struct {
	UserId string `json:"user_id" binding:"required"`
}
