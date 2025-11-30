package folder

import (
	"net/http"
	"ridash/repository"
	"ridash/utils/response"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// +----------------------------------------------+
// | GetFolder                                    |
// +----------------------------------------------+

// GetFolder godoc
// @Summary Get folder by ID
// @Description Retrieves a folder by its ID within a team
// @Tags folder
// @Accept json
// @Produce json
// @Param teamID path int true "Team ID"
// @Param id path int true "Folder ID"
// @Success 200 {object} response.SuccessResponse{data=models.Folder} "Folder retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid team ID or folder ID"
// @Failure 404 {object} response.ErrorResponse "Folder not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /teams/{teamID}/folders/{id} [get]
// @Security BearerAuth
func (h *FolderHandler) GetFolder(c echo.Context) error {
	teamIDStr := c.Param("teamID")
	teamID, err := strconv.ParseInt(teamIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team ID")
	}

	folderIDStr := c.Param("id")
	folderID, err := strconv.ParseInt(folderIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid folder ID")
	}

	tx, err := repository.StartTransaction(h.DB, c.Request().Context())
	if err != nil {
		zap.L().Error("Failed to begin transaction", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to begin transaction")
	}
	defer repository.DeferRollback(tx, c.Request().Context())

	folder, err := repository.GetFolderByIDAndTeamID(c.Request().Context(), tx, folderID, teamID)
	if err != nil {
		zap.L().Error("Failed to get folder", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get folder")
	}

	if folder == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Folder not found")
	}

	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		zap.L().Error("Failed to commit transaction", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit transaction")
	}

	return c.JSON(http.StatusOK, response.Success("Folder retrieved successfully", folder))
}
