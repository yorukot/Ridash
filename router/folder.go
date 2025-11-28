package router

import (
	"ridash/handler/folder"
	"ridash/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// FolderRouter handles all team-related routes
func FolderRouter(api *echo.Group, db *pgxpool.Pool) {
	folderHandler := &folder.FolderHandler{
		DB: db,
	}

	r := api.Group("/teams/:teamID/folders")
	r.POST("/", folderHandler.CreateFolder, middleware.AuthRequiredMiddleware)
	r.GET("/", folderHandler.GetFolders, middleware.AuthRequiredMiddleware)
	r.GET("/:id", folderHandler.GetFolder, middleware.AuthRequiredMiddleware)
	r.PUT("/:id", folderHandler.UpdateFolder, middleware.AuthRequiredMiddleware)
	r.DELETE("/:id", folderHandler.DeleteFolder, middleware.AuthRequiredMiddleware)
}
