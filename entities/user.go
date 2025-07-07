package entities

type User struct {
	ID       string   `gorm:"primaryKey"`
	Username string   `gorm:"type:varchar(100);unique;not null"`
	Hash     string   `gorm:"type:varchar(255);not null"`
	Role     UserRole `gorm:"type:varchar(10);not null"`
	Email    string   `gorm:"type:varchar(100);unique;not null"`
	Scopes   int64    `gorm:"not null;default:0"`
}

type UserRole string

const (
	Manager   UserRole = "manager"
	Developer UserRole = "developer"
)
