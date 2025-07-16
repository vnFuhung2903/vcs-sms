package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/vnFuhung2903/vcs-sms/entities"

	"gorm.io/gorm"
)

type IUserRepository interface {
	FindById(userId string) (*entities.User, error)
	FindByName(username string) (*entities.User, error)
	FindByEmail(email string) (*entities.User, error)
	Create(username, hash, email string, role entities.UserRole, scopes int64) (*entities.User, error)
	UpdatePassword(user *entities.User, hash string) error
	UpdateRole(user *entities.User, role entities.UserRole) error
	UpdateScope(user *entities.User, scopes int64) error
	Delete(userId string) error
	BeginTransaction(ctx context.Context) (*gorm.DB, error)
	WithTransaction(tx *gorm.DB) IUserRepository
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindById(userId string) (*entities.User, error) {
	var user entities.User
	res := r.db.First(&user, entities.User{ID: userId})
	if res.Error != nil {
		return nil, res.Error
	}
	return &user, nil
}

func (r *userRepository) FindByName(username string) (*entities.User, error) {
	var user entities.User
	res := r.db.First(&user, entities.User{Username: username})
	if res.Error != nil {
		return nil, res.Error
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*entities.User, error) {
	var user entities.User
	res := r.db.First(&user, entities.User{Email: email})
	if res.Error != nil {
		return nil, res.Error
	}
	return &user, nil
}

func (r *userRepository) Create(username, hash, email string, role entities.UserRole, scopes int64) (*entities.User, error) {
	newUser := &entities.User{
		ID:       uuid.New().String(),
		Username: username,
		Hash:     hash,
		Email:    email,
		Role:     role,
		Scopes:   scopes,
	}
	res := r.db.Create(newUser)
	if res.Error != nil {
		return nil, res.Error
	}
	return newUser, nil
}

func (r *userRepository) UpdatePassword(user *entities.User, hash string) error {
	res := r.db.Model(user).Update("hash", hash)
	return res.Error
}

func (r *userRepository) UpdateRole(user *entities.User, role entities.UserRole) error {
	res := r.db.Model(user).Update("role", role)
	return res.Error
}

func (r *userRepository) UpdateScope(user *entities.User, scopes int64) error {
	res := r.db.Model(user).Update("scopes", scopes)
	return res.Error
}

func (r *userRepository) Delete(userId string) error {
	res := r.db.Where("id = ?", userId).Delete(&entities.User{})
	return res.Error
}

func (r *userRepository) BeginTransaction(ctx context.Context) (*gorm.DB, error) {
	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return tx, nil
}

func (r *userRepository) WithTransaction(tx *gorm.DB) IUserRepository {
	return &userRepository{db: tx}
}
