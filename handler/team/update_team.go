package team

import (
	"encoding/json"
	"net/http"
	"ridash/repository"
	authutil "ridash/utils/auth"
	"ridash/utils/response"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// +----------------------------------------------+
// | UpdateTeam                                   |
// +----------------------------------------------+

type updateTeamRequest struct {
	Name string `json:"name" validate:"required,min=1,max=50" example:"Updated Team Name"`
}

// UpdateTeam godoc
// @Summary Update a team
// @Description Updates a team's information (only accessible by team owner)
// @Tags team
// @Accept json
// @Produce json
// @Param id path int true "Team ID"
// @Param request body updateTeamRequest true "Update team request"
// @Success 200 {object} response.SuccessResponse{data=models.Team} "Team updated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request body or team ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Only team owner can update the team"
// @Failure 404 {object} response.ErrorResponse "Team not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /teams/{id} [put]
// @Security BearerAuth
func (h *TeamHandler) UpdateTeam(c echo.Context) error {
	// Get the user ID from the context (set by auth middleware)
	userID, err := authutil.GetUserIDFromContext(c)
	if err != nil || userID == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	// Get the team ID from the URL parameter
	teamIDStr := c.Param("id")
	teamID, err := strconv.ParseInt(teamIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team ID")
	}

	// Decode the request body
	var updateTeamRequest updateTeamRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&updateTeamRequest); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate the request body
	if err := validator.New().Struct(updateTeamRequest); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Begin the transaction
	tx, err := repository.StartTransaction(h.DB, c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to begin transaction")
	}

	defer repository.DeferRollback(tx, c.Request().Context())

	// Get the team by ID
	team, err := repository.GetTeamByID(c.Request().Context(), tx, teamID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get team")
	}

	// If the team is not found, return an error
	if team == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Team not found")
	}

	// Check if the user is the owner of the team
	if team.OwnerID != *userID {
		return echo.NewHTTPError(http.StatusForbidden, "Only team owner can update the team")
	}

	// Update the team in the database
	if err = repository.UpdateTeam(c.Request().Context(), tx, teamID, updateTeamRequest.Name, time.Now()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update team")
	}

	// Update the team struct with new values
	team.Name = updateTeamRequest.Name
	team.UpdatedAt = time.Now()

	// Commit the transaction
	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit transaction")
	}

	// Respond with the success message and updated team data
	return c.JSON(http.StatusOK, response.Success("Team updated successfully", team))
}
