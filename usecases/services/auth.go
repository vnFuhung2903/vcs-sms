package services

import (
	"context"
	"net/mail"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/interfaces"
	"github.com/vnFuhung2903/vcs-sms/pkg/env"
	"github.com/vnFuhung2903/vcs-sms/pkg/logger"
	"github.com/vnFuhung2903/vcs-sms/usecases/repositories"
	"github.com/vnFuhung2903/vcs-sms/utils"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type IAuthService interface {
	Login(ctx context.Context, username, password string) (string, error)
	Register(username, password, email string, role entities.UserRole, scopes int64) (*entities.User, error)
	RefreshAccessToken(ctx context.Context, userId string) (string, error)
	UpdatePassword(ctx context.Context, userId, currentPassword, newPassword string) error
}

type authService struct {
	userRepo    repositories.IUserRepository
	redisClient interfaces.IRedisClient
	logger      logger.ILogger
	jwtSecret   []byte
}

func NewAuthService(userRepo repositories.IUserRepository, redisClient interfaces.IRedisClient, logger logger.ILogger, env env.AuthEnv) IAuthService {
	return &authService{
		userRepo:    userRepo,
		redisClient: redisClient,
		logger:      logger,
		jwtSecret:   []byte(env.JWTSecret),
	}
}

func (s *authService) Login(ctx context.Context, username, password string) (string, error) {
	var user *entities.User
	mail, err := mail.ParseAddress(username)
	if err != nil {
		user, err = s.userRepo.FindByName(username)
		if err != nil {
			s.logger.Error("failed to find user by username", zap.Error(err))
			return "", err
		}
	} else {
		user, err = s.userRepo.FindByEmail(mail.Address)
		if err != nil {
			s.logger.Error("failed to find user by email", zap.Error(err))
			return "", err
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(password)); err != nil {
		s.logger.Error("failed to validate password", zap.Error(err))
		return "", err
	}

	accessToken, err := s.generateAccessToken(user.ID, utils.HashMapToScopes(user.Scopes))
	if err != nil {
		s.logger.Error("failed to generate access token", zap.Error(err))
		return "", err
	}
	refreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		s.logger.Error("failed to generate refresh token", zap.Error(err))
		return "", err
	}

	if err := s.redisClient.Set(ctx, "refresh:"+user.ID, refreshToken, time.Hour*24*7); err != nil {
		s.logger.Error("failed to set refresh token in redis", zap.Error(err))
		return "", err
	}

	s.logger.Info("user logged in successfully")
	return accessToken, nil
}

func (s *authService) Register(username, password, email string, role entities.UserRole, scopes int64) (*entities.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", zap.Error(err))
		return nil, err
	}

	mail, err := mail.ParseAddress(email)
	if err != nil {
		s.logger.Error("failed to parse email", zap.Error(err))
		return nil, err
	}

	user, err := s.userRepo.Create(username, string(hash), mail.Address, role, scopes)
	if err != nil {
		s.logger.Error("failed to create user", zap.Error(err))
		return nil, err
	}

	s.logger.Info("new user registered successfully")
	return user, nil
}

func (s *authService) UpdatePassword(ctx context.Context, userId, currentPassword, newPassword string) error {
	user, err := s.userRepo.FindById(userId)
	if err != nil {
		s.logger.Error("failed to find user by id", zap.Error(err))
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(currentPassword)); err != nil {
		s.logger.Error("current password does not match", zap.Error(err))
		return err
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", zap.Error(err))
		return err
	}

	if err := s.userRepo.UpdatePassword(user, string(newHash)); err != nil {
		s.logger.Error("failed to update user's password", zap.Error(err))
		return err
	}

	if err := s.redisClient.Del(ctx, "refresh:"+user.ID); err != nil {
		s.logger.Error("failed to delete refresh token in redis", zap.Error(err))
		return err
	}

	s.logger.Info("user's password updated successfully")
	return nil
}

func (s *authService) RefreshAccessToken(ctx context.Context, userId string) (string, error) {
	refreshToken, err := s.redisClient.Get(ctx, "refresh:"+userId)
	if err != nil {
		s.logger.Error("failed to get refresh token from redis", zap.Error(err))
		return "", err
	}

	claims := &jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		s.logger.Error("invalid refresh token", zap.Error(err))
		return "", err
	}

	user, err := s.userRepo.FindById(userId)
	if err != nil {
		s.logger.Error("failed to find user by id", zap.Error(err))
		return "", err
	}

	accessToken, err := s.generateAccessToken(userId, utils.HashMapToScopes(user.Scopes))
	if err != nil {
		s.logger.Error("failed to generate new access token", zap.Error(err))
		return "", err
	}

	s.logger.Info("access token refreshed successfully")
	return accessToken, nil
}

func (s *authService) generateAccessToken(userId string, scope []string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userId,
		"scope": scope,
		"exp":   time.Now().Add(time.Minute * 15).Unix(),
		"iat":   time.Now().Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedAccessToken, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}
	return signedAccessToken, nil
}

func (s *authService) generateRefreshToken(userId string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userId,
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
		"iat": time.Now().Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedRefreshToken, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}
	return signedRefreshToken, nil
}
