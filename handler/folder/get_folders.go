package folder

import (
	"net/http"
	"ridash/repository"
	"ridash/utils/response"
	"strconv"

	"github.com/labstack/echo/v4"
)

// +----------------------------------------------+
// | GetFolders                                   |
// +----------------------------------------------+

// GetFolders godoc
// @Summary List folders for a team
// @Description Retrieves all folders belonging to a team
// @Tags folder
// @Accept json
// @Produce json
// @Param teamID path int true "Team ID"
// @Success 200 {object} response.SuccessResponse{data=[]models.Folder} "Folders retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid team ID"
// @Failure 404 {object} response.ErrorResponse "Team not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /teams/{teamID}/folders [get]
// @Security BearerAuth
func (h *FolderHandler) GetFolders(c echo.Context) error {
	teamIDStr := c.Param("teamID")
	teamID, err := strconv.ParseInt(teamIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team ID")
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

	folders, err := repository.GetFoldersByTeamID(c.Request().Context(), tx, teamID)
	if err != nil {
		return response.InternalServerError("Failed to get folders", err)
	}

	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		return response.InternalServerError("Failed to commit transaction", err)
	}

	return c.JSON(http.StatusOK, response.Success("Folders retrieved successfully", folders))
}
