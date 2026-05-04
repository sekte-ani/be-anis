package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"be-anis/helper"
	"be-anis/middleware"
	"be-anis/model"
	"be-anis/service"
)

type AuthController struct {
	authService service.AuthService
}

func NewAuthController(authService service.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

func (h *AuthController) RegisterRoutes(r *gin.Engine) {
	auth := r.Group("/api/auth")
	{
		auth.POST("/signup", h.SignUp)
		auth.POST("/signin", h.SignIn)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/forgot-password", h.ForgotPassword)
	}

	authed := r.Group("/api/auth")
	authed.Use(middleware.AuthRequired(h.authService))
	{
		authed.POST("/logout", h.Logout)
		authed.GET("/me", h.Me)
		authed.PUT("/profile", h.UpdateProfile)
		authed.PUT("/password", h.UpdatePassword)
	}
}

func (h *AuthController) SignUp(c *gin.Context) {
	var req model.SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Err(c, http.StatusBadRequest, "invalid request body", err)
		return
	}
	resp, err := h.authService.SignUp(req)
	if err != nil {
		helper.Err(c, http.StatusBadRequest, "signup failed", err)
		return
	}
	helper.OK(c, http.StatusCreated, "signup success", resp)
}

func (h *AuthController) SignIn(c *gin.Context) {
	var req model.SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Err(c, http.StatusBadRequest, "invalid request body", err)
		return
	}
	resp, err := h.authService.SignIn(req)
	if err != nil {
		helper.Err(c, http.StatusUnauthorized, "signin failed", err)
		return
	}
	helper.OK(c, http.StatusOK, "signin success", resp)
}

func (h *AuthController) Refresh(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Err(c, http.StatusBadRequest, "invalid request body", err)
		return
	}
	resp, err := h.authService.Refresh(req)
	if err != nil {
		helper.Err(c, http.StatusUnauthorized, "refresh failed", err)
		return
	}
	helper.OK(c, http.StatusOK, "token refreshed", resp)
}

func (h *AuthController) ForgotPassword(c *gin.Context) {
	var req model.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Err(c, http.StatusBadRequest, "invalid request body", err)
		return
	}
	if err := h.authService.ForgotPassword(req); err != nil {
		helper.Err(c, http.StatusBadRequest, "could not send recovery email", err)
		return
	}
	helper.OK(c, http.StatusOK, "recovery email sent", nil)
}

func (h *AuthController) Logout(c *gin.Context) {
	token := middleware.TokenFromContext(c)
	if err := h.authService.Logout(token); err != nil {
		helper.Err(c, http.StatusBadRequest, "logout failed", err)
		return
	}
	helper.OK(c, http.StatusOK, "logout success", nil)
}

func (h *AuthController) Me(c *gin.Context) {
	token := middleware.TokenFromContext(c)
	user, err := h.authService.GetCurrentUser(token)
	if err != nil {
		helper.Err(c, http.StatusUnauthorized, "could not load user", err)
		return
	}
	helper.OK(c, http.StatusOK, "current user", user)
}

func (h *AuthController) UpdateProfile(c *gin.Context) {
	var req model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Err(c, http.StatusBadRequest, "invalid request body", err)
		return
	}
	token := middleware.TokenFromContext(c)
	resp, err := h.authService.UpdateProfile(token, req)
	if err != nil {
		helper.Err(c, http.StatusBadRequest, "update profile failed", err)
		return
	}
	helper.OK(c, http.StatusOK, "profile updated", resp)
}

func (h *AuthController) UpdatePassword(c *gin.Context) {
	var req model.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Err(c, http.StatusBadRequest, "invalid request body", err)
		return
	}
	token := middleware.TokenFromContext(c)
	resp, err := h.authService.UpdatePassword(token, req)
	if err != nil {
		helper.Err(c, http.StatusBadRequest, "update password failed", err)
		return
	}
	helper.OK(c, http.StatusOK, "password updated", resp)
}
