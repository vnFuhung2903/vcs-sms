package services

import (
	"errors"
	"fmt"

	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/usecases/repositories"
	"github.com/vnFuhung2903/vcs-sms/utils/hashes"
)

type IUserService interface {
	Register(username, password, email string) (*entities.User, error)
	Login(username, password string) (*entities.User, error)
	UpdatePassword(userId, password string) error
	Delete(userId string) error
}

type userService struct {
	userRepo repositories.IUserRepository
}

func NewUserService(userRepo repositories.IUserRepository) IUserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) Register(username, password, email string) (*entities.User, error) {
	existing, _ := s.userRepo.FindByName(username)
	if existing != nil {
		return nil, fmt.Errorf("username already taken")
	}

	hash, err := hashes.HashPassword(password)
	if err != nil {
		return nil, err
	}

	return s.userRepo.Create(username, hash, email)
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

func (s *userService) Delete(userId string) error {
	return s.userRepo.Delete(userId)
}
