package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"be-anis/helper"
	"be-anis/middleware"
	"be-anis/model"
	"be-anis/service"
)

type MockController struct {
	mockService service.MockService
	authService service.AuthService
}

func NewMockController(mockService service.MockService, authService service.AuthService) *MockController {
	return &MockController{
		mockService: mockService,
		authService: authService,
	}
}

func (h *MockController) RegisterRoutes(r *gin.Engine) {
	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthRequired(h.authService))
	{
		admin.GET("/mocks", h.List)
		admin.GET("/mocks/:mock_id", h.Get)
		admin.POST("/mocks", h.Create)
		admin.PUT("/mocks/:mock_id", h.Update)
		admin.DELETE("/mocks/:mock_id", h.Delete)
		admin.POST("/mocks/generate-keywords", h.GenerateKeywords)
		admin.POST("/mocks/upload-image", h.UploadImage)
	}
}

func (h *MockController) List(c *gin.Context) {
	var query model.ListMocksQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		helper.Err(c, http.StatusBadRequest, "invalid query params", err)
		return
	}
	mocks, paginator, err := h.mockService.List(query)
	if err != nil {
		helper.Err(c, http.StatusInternalServerError, "failed to list mocks", err)
		return
	}
	helper.OKPaginated(c, http.StatusOK, "list mocks success", mocks, paginator)
}

func (h *MockController) Get(c *gin.Context) {
	mockID := c.Param("mock_id")
	mock, err := h.mockService.Get(mockID)
	if err != nil {
		helper.Err(c, http.StatusNotFound, "mock not found", err)
		return
	}
	helper.OK(c, http.StatusOK, "mock detail", mock)
}

func (h *MockController) Create(c *gin.Context) {
	var req model.CreateMockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Err(c, http.StatusBadRequest, "invalid request body", err)
		return
	}
	mock, err := h.mockService.Create(req)
	if err != nil {
		helper.Err(c, http.StatusInternalServerError, "failed to create mock", err)
		return
	}
	helper.OK(c, http.StatusCreated, "mock created", mock)
}

func (h *MockController) Update(c *gin.Context) {
	mockID := c.Param("mock_id")
	var req model.UpdateMockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Err(c, http.StatusBadRequest, "invalid request body", err)
		return
	}
	mock, err := h.mockService.Update(mockID, req)
	if err != nil {
		helper.Err(c, http.StatusInternalServerError, "failed to update mock", err)
		return
	}
	helper.OK(c, http.StatusOK, "mock updated", mock)
}

func (h *MockController) Delete(c *gin.Context) {
	mockID := c.Param("mock_id")
	if err := h.mockService.Delete(mockID); err != nil {
		helper.Err(c, http.StatusInternalServerError, "failed to delete mock", err)
		return
	}
	helper.OK(c, http.StatusOK, "mock deleted", nil)
}

func (h *MockController) UploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		helper.Err(c, http.StatusBadRequest, "image file is required (field: image)", err)
		return
	}
	url, err := h.mockService.UploadImage(file)
	if err != nil {
		helper.Err(c, http.StatusInternalServerError, "failed to upload image", err)
		return
	}
	helper.OK(c, http.StatusOK, "image uploaded", gin.H{"url": url})
}

func (h *MockController) GenerateKeywords(c *gin.Context) {
	var req model.GenerateKeywordsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helper.Err(c, http.StatusBadRequest, "invalid request body", err)
		return
	}
	resp, err := h.mockService.GenerateKeywords(req)
	if err != nil {
		helper.Err(c, http.StatusBadGateway, "failed to generate keywords", err)
		return
	}
	helper.OK(c, http.StatusOK, "keywords generated", resp)
}
