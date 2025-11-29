package folder

import (
	"net/http"
	"ridash/repository"
	authutil "ridash/utils/auth"
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
	userID, err := authutil.GetUserIDFromContext(c)
	if err != nil || userID == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
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
		return response.InternalServerError("Failed to begin transaction", err)
	}
	defer repository.DeferRollback(tx, c.Request().Context())

	team, err := repository.GetTeamByID(c.Request().Context(), tx, teamID)
	if err != nil {
		return response.InternalServerError("Failed to get team", err)
	}

	if team == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Team not found")
	}

	if team.OwnerID != *userID {
		return echo.NewHTTPError(http.StatusForbidden, "Only team owner can delete the folder")
	}

	folder, err := repository.GetFolderByIDAndTeamID(c.Request().Context(), tx, folderID, teamID)
	if err != nil {
		return response.InternalServerError("Failed to get folder", err)
	}

	if folder == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Folder not found")
	}

	if err := repository.DeleteFolder(c.Request().Context(), tx, folderID, teamID); err != nil {
		return response.InternalServerError("Failed to delete folder", err)
	}

	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		return response.InternalServerError("Failed to commit transaction", err)
	}

	return c.JSON(http.StatusOK, response.SuccessMessage("Folder deleted successfully"))
}
