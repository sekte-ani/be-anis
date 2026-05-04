package model

import "github.com/supabase-community/gotrue-go/types"

type SignUpRequest struct {
	Email    string                 `json:"email" binding:"required,email"`
	Password string                 `json:"password" binding:"required,min=6"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

type SignInRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type UpdatePasswordRequest struct {
	Password string `json:"password" binding:"required,min=6"`
}

type UpdateProfileRequest struct {
	Email string                 `json:"email,omitempty"`
	Phone string                 `json:"phone,omitempty"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

type AuthResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	TokenType    string      `json:"token_type"`
	ExpiresIn    int         `json:"expires_in"`
	ExpiresAt    int64       `json:"expires_at"`
	User         types.User  `json:"user"`
}

func ToAuthResponse(s types.Session) AuthResponse {
	return AuthResponse{
		AccessToken:  s.AccessToken,
		RefreshToken: s.RefreshToken,
		TokenType:    s.TokenType,
		ExpiresIn:    s.ExpiresIn,
		ExpiresAt:    s.ExpiresAt,
		User:         s.User,
	}
}
