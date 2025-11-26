package response

// ErrorResponse represents an error response
type ErrorResponse struct {
	Message string `json:"message" example:"Invalid request body"`
}

// SuccessResponse represents a success response with optional data
type SuccessResponse struct {
	Message string      `json:"message" example:"Operation successful"`
	Data    any `json:"data,omitempty" swaggertype:"object"`
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
