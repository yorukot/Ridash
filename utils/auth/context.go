package auth

import (
	"strconv"

	"ridash/middleware"

	"github.com/labstack/echo/v4"
)

// GetUserIDFromContext returns the user ID if present in context.
func GetUserIDFromContext(c echo.Context) (*int64, error) {
	userIDStr, ok := c.Get(string(middleware.UserIDKey)).(string)
	if !ok || userIDStr == "" {
		return nil, nil
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return nil, err
	}

	return &userID, nil
}
