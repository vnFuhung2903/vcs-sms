package entities

import (
	"fmt"
	"time"
)

type Container struct {
	ContainerId   string          `gorm:"primaryKey"`
	Status        ContainerStatus `gorm:"type:varchar(10)"`
	CreatedAt     time.Time       `gorm:"autoCreateTime"`
	UpdatedAt     time.Time       `gorm:"autoUpdateTime"`
	ContainerName string          `gorm:"unique"`
	Ipv4          string          `gorm:"unique"`
}

type ContainerFilter struct {
	ContainerId   string
	Status        ContainerStatus
	ContainerName string
	Ipv4          string
}

type ContainerSort struct {
	Field string
	Sort  SortOrder
}

var sortField = map[string]bool{
	"container_id":   true,
	"container_name": true,
	"status":         true,
	"ipv4":           true,
	"created_at":     true,
	"updated_at":     true,
}

func ValidateSort(sort ContainerSort) error {
	if !sortField[sort.Field] {
		return fmt.Errorf("invalid sort field")
	}
	if sort.Sort != Asc && sort.Sort != Dsc {
		return fmt.Errorf("invalid sort order: %s", sort.Sort)
	}
	return nil
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
