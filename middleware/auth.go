package middleware

import (
	"net/http"
	"ridash/utils/config"
	"ridash/utils/encrypt"
	"strings"

	"github.com/labstack/echo/v4"
)

// authMiddlewareLogic is the logic for the auth middleware
func authMiddlewareLogic(token string) (*encrypt.AccessTokenClaims, error) {
	token = strings.TrimPrefix(token, "Bearer ")

	JWTSecret := encrypt.JWTSecret{
		Secret: config.Env().JWTSecretKey,
	}

	valid, claims, err := JWTSecret.ValidateAccessTokenAndGetClaims(token)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	if !valid {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
	}

	return &claims, nil
}

// AuthRequiredMiddleware is the middleware for the auth required
func AuthRequiredMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		}

		claims, err := authMiddlewareLogic(token)
		if err != nil {
			return err
		}

		c.Set(string(UserIDKey), claims.Subject)
		return next(c)
	}
}

// AuthOptionalMiddleware is the middleware for the auth optional
func AuthOptionalMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			return next(c)
		}

		claims, err := authMiddlewareLogic(token)
		if err != nil {
			// For optional auth, continue even if token is invalid
			return next(c)
		}

		c.Set(string(UserIDKey), claims.Subject)
		return next(c)
	}
}
