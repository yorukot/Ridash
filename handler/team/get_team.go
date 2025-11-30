package team

import (
	"net/http"
	"ridash/repository"
	"ridash/utils/response"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// +----------------------------------------------+
// | GetTeam                                      |
// +----------------------------------------------+

// GetTeam godoc
// @Summary Get team by ID
// @Description Retrieves a team by its ID
// @Tags team
// @Accept json
// @Produce json
// @Param id path int true "Team ID"
// @Success 200 {object} response.SuccessResponse{data=models.Team} "Team retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid team ID"
// @Failure 404 {object} response.ErrorResponse "Team not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /teams/{id} [get]
func (h *TeamHandler) GetTeam(c echo.Context) error {
	// Get the team ID from the URL parameter
	teamIDStr := c.Param("id")
	teamID, err := strconv.ParseInt(teamIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team ID")
	}

	// Begin the transaction
	tx, err := repository.StartTransaction(h.DB, c.Request().Context())
	if err != nil {
		zap.L().Error("Failed to begin transaction", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to begin transaction")
	}

	defer repository.DeferRollback(tx, c.Request().Context())

	// Get the team by ID
	team, err := repository.GetTeamByID(c.Request().Context(), tx, teamID)
	if err != nil {
		zap.L().Error("Failed to get team", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get team")
	}

	// If the team is not found, return an error
	if team == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Team not found")
	}

	// Commit the transaction
	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		zap.L().Error("Failed to commit transaction", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit transaction")
	}

	// Respond with the success message and team data
	return c.JSON(http.StatusOK, response.Success("Team retrieved successfully", team))
}
