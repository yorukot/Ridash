package router

import (
	"ridash/handler/document"
	"ridash/middleware"
	"ridash/utils/config"
	"ridash/utils/docmanager"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Auth router going to route register signin etc
func DocumentRouter(api *echo.Group, db *pgxpool.Pool) {
	docManager, err := docmanager.NewClient(config.Env().DocManagerBaseURL, config.Env().DocManagerAPIToken)
	if err != nil {
		zap.L().Fatal("Failed to initialize document manager client", zap.Error(err))
	}

	documentHandler := &document.DocumentHandler{
		DB:         db,
		DocManager: docManager,
	}

	// Publicly readable endpoints (respect document permission checks in handlers)
	readable := api.Group("/documents", middleware.AuthOptionalMiddleware)
	readable.GET("", documentHandler.ListDocuments)
	readable.GET("/:id", documentHandler.GetDocument)

	// Authenticated endpoints for owners/collaborators
	protected := api.Group("/documents", middleware.AuthRequiredMiddleware)
	protected.POST("", documentHandler.CreateDocument)
	protected.PUT("/:id", documentHandler.UpdateDocument)
	protected.DELETE("/:id", documentHandler.DeleteDocument)
	protected.GET("/:id/socket", documentHandler.ProxyDocumentWebsocket)

	shares := protected.Group("/:id/shares")
	shares.GET("", documentHandler.ListShares)
	shares.POST("", documentHandler.CreateShare)
	shares.PUT("/:shareID", documentHandler.UpdateShare)
	shares.DELETE("/:shareID", documentHandler.DeleteShare)
}
