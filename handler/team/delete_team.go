package team

import (
	"net/http"
	"ridash/repository"
	authutil "ridash/utils/auth"
	"ridash/utils/response"
	"strconv"

	"github.com/labstack/echo/v4"
)

// +----------------------------------------------+
// | DeleteTeam                                   |
// +----------------------------------------------+

// DeleteTeam godoc
// @Summary Delete a team
// @Description Deletes a team and its memberships (owner only)
// @Tags team
// @Accept json
// @Produce json
// @Param id path int true "Team ID"
// @Success 200 {object} response.SuccessResponse "Team deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid team ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Only team owner can delete the team"
// @Failure 404 {object} response.ErrorResponse "Team not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /teams/{id} [delete]
// @Security BearerAuth
func (h *TeamHandler) DeleteTeam(c echo.Context) error {
	userID, err := authutil.GetUserIDFromContext(c)
	if err != nil || userID == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	teamIDStr := c.Param("id")
	teamID, err := strconv.ParseInt(teamIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team ID")
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

	if team.OwnerID != *userID {
		return echo.NewHTTPError(http.StatusForbidden, "Only team owner can delete the team")
	}

	if err := repository.DeleteTeamMembersByTeamID(c.Request().Context(), tx, teamID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete team members")
	}

	if err := repository.DeleteTeam(c.Request().Context(), tx, teamID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete team")
	}

	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit transaction")
	}

	return c.JSON(http.StatusOK, response.SuccessMessage("Team deleted successfully"))
}
