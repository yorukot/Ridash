package router

import (
	"ridash/handler/auth"
	"ridash/middleware"
	"ridash/utils/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// AuthRouter wires authentication and OAuth routes
func AuthRouter(api *echo.Group, db *pgxpool.Pool) {
	oauthConfig, err := config.GetOAuthConfig()
	if err != nil {
		zap.L().Fatal("Failed to initialize OAuth config", zap.Error(err))
	}

	authHandler := &auth.AuthHandler{
		DB:          db,
		OAuthConfig: oauthConfig,
	}

	r := api.Group("/auth")
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.POST("/refresh", authHandler.RefreshToken)

	// OAuth routes allow optional auth for linking existing accounts
	oauth := r.Group("/oauth", middleware.AuthOptionalMiddleware)
	oauth.GET("/:provider", authHandler.OAuthEntry)
	oauth.GET("/:provider/callback", authHandler.OAuthCallback)
}
