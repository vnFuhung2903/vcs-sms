package entities

import (
	"time"
)

type Server struct {
	ServerId   string       `gorm:"primaryKey"`
	Status     ServerStatus `gorm:"type:varchar(10)"`
	CreatedAt  time.Time    `gorm:"autoCreateTime"`
	UpdatedAt  time.Time    `gorm:"autoUpdateTime"`
	ServerName string       `gorm:"unique"`
	Ipv4       string       `gorm:"unique"`
}

type ServerFilter struct {
	ServerId   string
	Status     ServerStatus
	ServerName string
	Ipv4       string
}

type ServerSort struct {
	Field string
	Sort  SortOrder
}

type SortOrder string

const (
	Asc SortOrder = "ASC"
	Dsc SortOrder = "DESC"
)

type ServerStatus string

const (
	ServerOn  ServerStatus = "ON"
	ServerOff ServerStatus = "OFF"
)
