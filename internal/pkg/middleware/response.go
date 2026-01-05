package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// APIResponse represents standard API response format
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *MetaInfo   `json:"meta"`
}

// ErrorInfo represents error information
type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// MetaInfo represents metadata
type MetaInfo struct {
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// SuccessResponse sends a successful API response
func SuccessResponse(c *gin.Context, statusCode int, data interface{}) {
	requestID, _ := c.Get("request_id")

	c.JSON(statusCode, APIResponse{
		Success: true,
		Data:    data,
		Meta: &MetaInfo{
			RequestID: requestID.(string),
			Timestamp: time.Now().UTC(),
			Version:   "v1",
		},
	})
}

// ErrorResponse sends an error API response
func ErrorResponse(c *gin.Context, statusCode int, code string, message string, details interface{}) {
	requestID, _ := c.Get("request_id")

	c.JSON(statusCode, APIResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
		Meta: &MetaInfo{
			RequestID: requestID.(string),
			Timestamp: time.Now().UTC(),
			Version:   "v1",
		},
	})
}
