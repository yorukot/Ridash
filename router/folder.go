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

	r := api.Group("/teams/:teamID/folders", middleware.AuthRequiredMiddleware)
	r.POST("/", folderHandler.CreateFolder)
	r.GET("/", folderHandler.GetFolders)
	r.GET("/:id", folderHandler.GetFolder)
	r.PUT("/:id", folderHandler.UpdateFolder)
	r.DELETE("/:id", folderHandler.DeleteFolder)
}
