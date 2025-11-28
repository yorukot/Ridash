package router

import (
	"ridash/handler/document"
	"ridash/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// Auth router going to route register signin etc
func DocumentRouter(api *echo.Group, db *pgxpool.Pool) {
	documentHandler := &document.DocumentHandler{
		DB: db,
	}

	// Publicly readable endpoints (respect document permission checks in handlers)
	readable := api.Group("/documents", middleware.AuthOptionalMiddleware)
	readable.GET("/", documentHandler.ListDocuments)
	readable.GET("/:id", documentHandler.GetDocument)

	// Authenticated endpoints for owners/collaborators
	protected := api.Group("/documents", middleware.AuthRequiredMiddleware)
	protected.POST("/", documentHandler.CreateDocument)
	protected.PUT("/:id", documentHandler.UpdateDocument)
	protected.DELETE("/:id", documentHandler.DeleteDocument)

	shares := protected.Group("/:id/shares")
	shares.GET("/", documentHandler.ListShares)
	shares.POST("/", documentHandler.CreateShare)
	shares.PUT("/:shareID", documentHandler.UpdateShare)
	shares.DELETE("/:shareID", documentHandler.DeleteShare)
}
