package dto

import (
	"github.com/vnFuhung2903/vcs-sms/entities"
)

type CreateRequest struct {
	ContainerName string `json:"container_name"`
	ImageName     string `json:"image_name"`
}

type ViewResponse struct {
	Data  []*entities.Container `json:"data"`
	Total int64                 `json:"total"`
}

type ImportResponse struct {
	SuccessCount      int      `json:"success_count"`
	SuccessContainers []string `json:"success_containers"`
	FailedCount       int      `json:"failed_count"`
	FailedContainers  []string `json:"failed_containers"`
}

type ContainerUpdate struct {
	Status entities.ContainerStatus `json:"status"`
}

type ContainerFilter struct {
	ContainerId   string                   `form:"container_id"`
	Status        entities.ContainerStatus `form:"status" binding:"omitempty,oneof=on off starting stopped error"`
	ContainerName string                   `form:"container_name"`
	Ipv4          string                   `form:"ipv4"`
}

type ContainerSort struct {
	Field string    `form:"field" binding:"omitempty,oneof=container_id container_name status ipv4 created_at updated_at"`
	Order SortOrder `form:"order" binding:"omitempty,oneof=asc desc"`
}

var SortField = map[string]bool{
	"container_id":   true,
	"container_name": true,
	"status":         true,
	"ipv4":           true,
	"created_at":     true,
	"updated_at":     true,
}

type SortOrder string

const (
	Asc SortOrder = "asc"
	Dsc SortOrder = "desc"
)
