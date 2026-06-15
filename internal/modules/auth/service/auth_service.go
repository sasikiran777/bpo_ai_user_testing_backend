package service

import (
	"context"
	"errors"

	"ai_testing/internal/modules/auth/dto"
	"ai_testing/internal/modules/auth/utils"
	"ai_testing/internal/modules/users/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.Repository
}

func New(
	repo *repository.Repository,
) *AuthService {
	return &AuthService{
		userRepo: repo,
	}
}

func (s *AuthService) Login(
	ctx context.Context,
	payload dto.LoginRequest,
) (*dto.LoginResponse, error) {

	user, err := s.userRepo.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := utils.GenerateToken(payload.Email)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		Token:     token,
		FirstName: user.FirstName,
	}, nil
}

func (s *AuthService) Register(
	ctx context.Context,
	payload dto.RegisterRequest,
) (*dto.LoginResponse, error) {
	user := payload.ToUserModel()
	user.Password = utils.HashPassword(user.Password)
	user, err := s.userRepo.Register(ctx, user)
	if err != nil {
		return nil, err
	}
	token, err := utils.GenerateToken(user.Email)
	if err != nil {
		return nil, err
	}
	return &dto.LoginResponse{
		Token:     token,
		FirstName: user.FirstName,
	}, nil
}
