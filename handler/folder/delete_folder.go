package folder

import (
	"net/http"
	"ridash/middleware"
	"ridash/repository"
	"ridash/utils/response"
	"strconv"

	"github.com/labstack/echo/v4"
)

// +----------------------------------------------+
// | DeleteFolder                                 |
// +----------------------------------------------+

// DeleteFolder godoc
// @Summary Delete a folder
// @Description Deletes a folder within a team (owner only)
// @Tags folder
// @Accept json
// @Produce json
// @Param teamID path int true "Team ID"
// @Param id path int true "Folder ID"
// @Success 200 {object} response.SuccessResponse "Folder deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid team ID or folder ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Only team owner can delete the folder"
// @Failure 404 {object} response.ErrorResponse "Team or folder not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /teams/{teamID}/folders/{id} [delete]
// @Security BearerAuth
func (h *FolderHandler) DeleteFolder(c echo.Context) error {
	userIDStr, ok := c.Get(string(middleware.UserIDKey)).(string)
	if !ok || userIDStr == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID")
	}

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
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to begin transaction")
	}
	defer repository.DeferRollback(tx, c.Request().Context())

	team, err := repository.GetTeamByID(c.Request().Context(), tx, teamID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get team")
	}

	if team == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Team not found")
	}

	if team.OwnerID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "Only team owner can delete the folder")
	}

	folder, err := repository.GetFolderByIDAndTeamID(c.Request().Context(), tx, folderID, teamID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get folder")
	}

	if folder == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Folder not found")
	}

	if err := repository.DeleteFolder(c.Request().Context(), tx, folderID, teamID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete folder")
	}

	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit transaction")
	}

	return c.JSON(http.StatusOK, response.SuccessMessage("Folder deleted successfully"))
}
