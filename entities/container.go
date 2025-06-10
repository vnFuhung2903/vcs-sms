package entities

import (
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

type SortOrder string

const (
	Asc SortOrder = "ASC"
	Dsc SortOrder = "DESC"
)

type ContainerStatus string

const (
	ContainerOn  ContainerStatus = "ON"
	ContainerOff ContainerStatus = "OFF"
)
