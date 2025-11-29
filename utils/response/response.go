package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Message string `json:"message" example:"Invalid request body"`
	Detail  string `json:"detail,omitempty" example:"duplicate key value violates unique constraint \"teams_name_key\""`
}

// SuccessResponse represents a success response with optional data
type SuccessResponse struct {
	Message string `json:"message" example:"Operation successful"`
	Data    any    `json:"data,omitempty" swaggertype:"object"`
}

// Success sends a success response with a message and optional data
func Success(message string, data any) SuccessResponse {
	return SuccessResponse{
		Message: message,
		Data:    data,
	}
}

// SuccessMessage sends a success response with only a message
func SuccessMessage(message string) SuccessResponse {
	return SuccessResponse{
		Message: message,
	}
}

// InternalServerError logs the error and returns a standardized HTTP 500 error.
func InternalServerError(message string, err error) *echo.HTTPError {
	if err != nil {
		zap.L().Error(message, zap.Error(err))
	} else {
		zap.L().Error(message)
	}

	httpErr := echo.NewHTTPError(http.StatusInternalServerError, message)
	if err != nil {
		httpErr.Internal = err
	}

	return httpErr
}
