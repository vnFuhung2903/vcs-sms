package entities

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID       string    `gorm:"primaryKey"`
	Username string    `gorm:"type:varchar(100);unique;not null"`
	Hash     string    `gorm:"type:varchar(255);not null"`
	Scope    UserScope `gorm:"type:varchar(10);not null"`
	Email    string    `gorm:"type:varchar(100);unique;not null"`
}

type UserScope string

const (
	Admin UserScope = "admin"
	Read  UserScope = "read"
	Write UserScope = "write"
)
