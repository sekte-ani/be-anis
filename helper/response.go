package helper

import "github.com/gin-gonic/gin"

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func OK(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Err(c *gin.Context, status int, message string, err error) {
	resp := Response{
		Success: false,
		Message: message,
	}
	if err != nil {
		resp.Error = err.Error()
	}
	c.AbortWithStatusJSON(status, resp)
}
