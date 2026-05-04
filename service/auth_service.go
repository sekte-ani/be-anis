package service

import (
	"errors"

	"github.com/supabase-community/gotrue-go/types"

	"be-anis/model"
	"be-anis/repository"
)

type AuthService interface {
	SignUp(req model.SignUpRequest) (*types.SignupResponse, error)
	SignIn(req model.SignInRequest) (model.AuthResponse, error)
	Refresh(req model.RefreshTokenRequest) (model.AuthResponse, error)
	Logout(accessToken string) error
	ForgotPassword(req model.ForgotPasswordRequest) error
	GetCurrentUser(accessToken string) (*types.UserResponse, error)
	UpdateProfile(accessToken string, req model.UpdateProfileRequest) (*types.UpdateUserResponse, error)
	UpdatePassword(accessToken string, req model.UpdatePasswordRequest) (*types.UpdateUserResponse, error)
}

type authService struct {
	repo repository.AuthRepository
}

func NewAuthService(repo repository.AuthRepository) AuthService {
	return &authService{repo: repo}
}

func (s *authService) SignUp(req model.SignUpRequest) (*types.SignupResponse, error) {
	return s.repo.SignUp(types.SignupRequest{
		Email:    req.Email,
		Password: req.Password,
		Data:     req.Data,
	})
}

func (s *authService) SignIn(req model.SignInRequest) (model.AuthResponse, error) {
	session, err := s.repo.SignInWithPassword(req.Email, req.Password)
	if err != nil {
		return model.AuthResponse{}, err
	}
	return model.ToAuthResponse(session), nil
}

func (s *authService) Refresh(req model.RefreshTokenRequest) (model.AuthResponse, error) {
	session, err := s.repo.RefreshToken(req.RefreshToken)
	if err != nil {
		return model.AuthResponse{}, err
	}
	return model.ToAuthResponse(session), nil
}

func (s *authService) Logout(accessToken string) error {
	if accessToken == "" {
		return errors.New("missing access token")
	}
	return s.repo.Logout(accessToken)
}

func (s *authService) ForgotPassword(req model.ForgotPasswordRequest) error {
	return s.repo.Recover(req.Email)
}

func (s *authService) GetCurrentUser(accessToken string) (*types.UserResponse, error) {
	return s.repo.GetUser(accessToken)
}

func (s *authService) UpdateProfile(accessToken string, req model.UpdateProfileRequest) (*types.UpdateUserResponse, error) {
	return s.repo.UpdateUser(accessToken, types.UpdateUserRequest{
		Email: req.Email,
		Phone: req.Phone,
		Data:  req.Data,
	})
}

func (s *authService) UpdatePassword(accessToken string, req model.UpdatePasswordRequest) (*types.UpdateUserResponse, error) {
	pwd := req.Password
	return s.repo.UpdateUser(accessToken, types.UpdateUserRequest{
		Password: &pwd,
	})
}
