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

type ContainerUpdate struct {
	Status entities.ContainerStatus `json:"status"`
}

type ContainerFilter struct {
	ContainerId   string                   `form:"container_id"`
	Status        entities.ContainerStatus `form:"status"`
	ContainerName string                   `form:"container_name"`
	Ipv4          string                   `form:"ipv4"`
}

type ContainerSort struct {
	Field string    `form:"field"`
	Sort  SortOrder `form:"order"`
}

var sortField = map[string]bool{
	"container_id":   true,
	"container_name": true,
	"status":         true,
	"ipv4":           true,
	"created_at":     true,
	"updated_at":     true,
}

func StandardizeSort(sort ContainerSort) ContainerSort {
	standardizedSort := sort
	if !sortField[sort.Field] {
		standardizedSort.Field = "container_id"
	}
	if sort.Sort != Asc {
		standardizedSort.Sort = Dsc
	}
	return standardizedSort
}

type SortOrder string

const (
	Asc SortOrder = "asc"
	Dsc SortOrder = "desc"
)
