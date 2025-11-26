package router

import (
	"ridash/handler/auth"
	"ridash/middleware"
	"ridash/utils/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// Auth router going to route register signin etc
func AuthRouter(api *echo.Group, db *pgxpool.Pool) {
	oauthConfig, err := config.GetOAuthConfig()
	if err != nil {
		panic("Failed to initialize OAuth config: " + err.Error())
	}

	authHandler := &auth.AuthHandler{
		DB:          db,
		OAuthConfig: oauthConfig,
	}
	r := api.Group("/auth")

	r.GET("/oauth/:provider", authHandler.OAuthEntry, middleware.AuthOptionalMiddleware)
	r.GET("/oauth/:provider/callback", authHandler.OAuthCallback)

	r.POST("/register", authHandler.Register)
}
