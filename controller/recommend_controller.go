package controller

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"be-anis/helper"
	"be-anis/model"
	"be-anis/service"
)

type RecommendController struct {
	recommendService service.RecommendService
}

func NewRecommendController(recommendService service.RecommendService) *RecommendController {
	return &RecommendController{recommendService: recommendService}
}

func (h *RecommendController) RegisterRoutes(r *gin.Engine) {
	r.POST("/recommend", h.Recommend)
	r.POST("/reccomend", h.Recommend)
	log.Printf("[recommend][route] registered method=POST path=/recommend handler=Recommend")
	log.Printf("[recommend][route] registered method=POST path=/reccomend handler=Recommend (legacy typo alias)")
}

func (h *RecommendController) Recommend(c *gin.Context) {
	startedAt := time.Now()
	method := c.Request.Method
	path := c.Request.URL.Path
	route := c.FullPath()
	if route == "" {
		route = path
	}
	clientIP := c.ClientIP()

	log.Printf("[recommend][http] request_received method=%s path=%s route=%s client_ip=%s", method, path, route, clientIP)

	var req model.RecommendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf(
			"[recommend][http] request_invalid method=%s path=%s route=%s status=%d duration=%s err=%v",
			method, path, route, http.StatusBadRequest, time.Since(startedAt), err,
		)
		helper.Err(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	cleanInput := compactSpaces(req.Input)
	log.Printf(
		"[recommend][http] request_bound method=%s path=%s route=%s input_chars=%d input_preview=%q",
		method, path, route, len([]rune(cleanInput)), previewText(cleanInput, 160),
	)

	resp, err := h.recommendService.Recommend(req)
	if err != nil {
		log.Printf(
			"[recommend][http] request_failed method=%s path=%s route=%s status=%d duration=%s err=%v",
			method, path, route, http.StatusInternalServerError, time.Since(startedAt), err,
		)
		helper.Err(c, http.StatusInternalServerError, "failed to generate recommendation", err)
		return
	}

	log.Printf(
		"[recommend][http] request_succeeded method=%s path=%s route=%s status=%d duration_ms=%d sektor=%q mock_count=%d fitur_chars=%d",
		method, path, route, http.StatusOK, time.Since(startedAt).Milliseconds(), resp.SektorTerdeteksi, len(resp.MockReferences), len([]rune(resp.RekomendasiFitur)),
	)

	c.JSON(http.StatusOK, resp)
}

func compactSpaces(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

func previewText(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}
