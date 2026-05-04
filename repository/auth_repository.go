package repository

import (
	"github.com/supabase-community/gotrue-go/types"

	"be-anis/config"
)

type AuthRepository interface {
	SignUp(req types.SignupRequest) (*types.SignupResponse, error)
	SignInWithPassword(email, password string) (types.Session, error)
	RefreshToken(refreshToken string) (types.Session, error)
	Logout(accessToken string) error
	Recover(email string) error
	GetUser(accessToken string) (*types.UserResponse, error)
	UpdateUser(accessToken string, req types.UpdateUserRequest) (*types.UpdateUserResponse, error)
}

type authRepository struct {
	clients *config.SupabaseClients
}

func NewAuthRepository(clients *config.SupabaseClients) AuthRepository {
	return &authRepository{clients: clients}
}

func (r *authRepository) SignUp(req types.SignupRequest) (*types.SignupResponse, error) {
	return r.clients.Public.Auth.Signup(req)
}

func (r *authRepository) SignInWithPassword(email, password string) (types.Session, error) {
	return r.clients.Public.SignInWithEmailPassword(email, password)
}

func (r *authRepository) RefreshToken(refreshToken string) (types.Session, error) {
	return r.clients.Public.RefreshToken(refreshToken)
}

func (r *authRepository) Logout(accessToken string) error {
	return r.clients.Public.Auth.WithToken(accessToken).Logout()
}

func (r *authRepository) Recover(email string) error {
	return r.clients.Public.Auth.Recover(types.RecoverRequest{Email: email})
}

func (r *authRepository) GetUser(accessToken string) (*types.UserResponse, error) {
	return r.clients.Public.Auth.WithToken(accessToken).GetUser()
}

func (r *authRepository) UpdateUser(accessToken string, req types.UpdateUserRequest) (*types.UpdateUserResponse, error) {
	return r.clients.Public.Auth.WithToken(accessToken).UpdateUser(req)
}
