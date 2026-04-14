package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SuccessResponse struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Data    *any   `json:"data"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    data,
	})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Data:    data,
	})
}

func Error(c *gin.Context, code int, msg string, data ...any) {
	var respData *any
	if len(data) > 0 && data[0] != nil {
		respData = &data[0]
	}

	c.JSON(code, ErrorResponse{
		Success: false,
		Error:   msg,
		Data:    respData,
	})
}
