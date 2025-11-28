package router

import (
	"ridash/handler/team"
	"ridash/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// TeamRouter handles all team-related routes
func TeamRouter(api *echo.Group, db *pgxpool.Pool) {
	teamHandler := &team.TeamHandler{
		DB: db,
	}

	r := api.Group("/teams")
	r.POST("/", teamHandler.CreateTeam, middleware.AuthRequiredMiddleware)
}
