package entities

import (
	"time"
)

type Container struct {
	ContainerId   string          `gorm:"primaryKey"`
	Status        ContainerStatus `gorm:"type:varchar(10);not null"`
	CreatedAt     time.Time       `gorm:"autoCreateTime"`
	UpdatedAt     time.Time       `gorm:"autoUpdateTime"`
	ContainerName string          `gorm:"unique;not null"`
	Ipv4          string          `gorm:"unique;not null"`
}

type ContainerUpdate struct {
	Status ContainerStatus `json:"status"`
}

type ContainerFilter struct {
	ContainerId   string          `form:"container_id"`
	Status        ContainerStatus `form:"status"`
	ContainerName string          `form:"container_name"`
	Ipv4          string          `form:"ipv4"`
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
	var standardizedSort ContainerSort
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

type ContainerStatus string

const (
	ContainerOn  ContainerStatus = "ON"
	ContainerOff ContainerStatus = "OFF"
)
