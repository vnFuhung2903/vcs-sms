package services

import (
	"errors"

	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/usecases/repositories"
	"github.com/vnFuhung2903/vcs-sms/utils/hashes"
)

type IUserService interface {
	Register(username, password, email string, role entities.UserRole, scopes int64) (*entities.User, error)
	Login(username, password string) (*entities.User, error)
	UpdatePassword(userId, password string) error
	UpdateRole(userId string, role entities.UserRole) error
	UpdateScope(userId string, scopes int64) error
	Delete(userId string) error
}

type userService struct {
	userRepo repositories.IUserRepository
}

func NewUserService(userRepo repositories.IUserRepository) IUserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) Register(username, password, email string, role entities.UserRole, scopes int64) (*entities.User, error) {
	existing, _ := s.userRepo.FindByName(username)
	if existing != nil {
		return nil, errors.New("username already taken")
	}

	hash, err := hashes.HashPassword(password)
	if err != nil {
		return nil, err
	}

	return s.userRepo.Create(username, hash, email, role, scopes)
}

func (s *userService) Login(username, password string) (*entities.User, error) {
	user, err := s.userRepo.FindByName(username)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if err := hashes.ValidatePassword(password, user.Hash); err != nil {
		return nil, errors.New("invalid password")
	}

	return user, nil
}

func (s *userService) UpdatePassword(userId, password string) error {
	user, err := s.userRepo.FindById(userId)
	if err != nil {
		return err
	}

	hash, err := hashes.HashPassword(password)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(user, hash)
}

func (s *userService) UpdateRole(userId string, role entities.UserRole) error {
	user, err := s.userRepo.FindById(userId)
	if err != nil {
		return err
	}
	return s.userRepo.UpdateRole(user, role)
}

func (s *userService) UpdateScope(userId string, scopes int64) error {
	user, err := s.userRepo.FindById(userId)
	if err != nil {
		return err
	}
	return s.userRepo.UpdateScope(user, scopes)
}

func (s *userService) Delete(userId string) error {
	return s.userRepo.Delete(userId)
}
