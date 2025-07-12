package dto

import (
	"github.com/vnFuhung2903/vcs-sms/entities"
)

type CreateRequest struct {
	ContainerId   string                   `json:"container_id"`
	ContainerName string                   `json:"container_name"`
	Status        entities.ContainerStatus `json:"status"`
	IPv4          string                   `json:"ipv4"`
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
