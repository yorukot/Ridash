package team

import (
	"encoding/json"
	"net/http"
	"ridash/models"
	"ridash/repository"
	authutil "ridash/utils/auth"
	"ridash/utils/id"
	"ridash/utils/response"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// +----------------------------------------------+
// | CreateTeam                                   |
// +----------------------------------------------+

type createTeamRequest struct {
	Name string `json:"name" validate:"required,min=1,max=50" example:"My Team"`
}

// CreateTeam godoc
// @Summary Create a new team
// @Description Creates a new team with the authenticated user as the owner
// @Tags team
// @Accept json
// @Produce json
// @Param request body createTeamRequest true "Create team request"
// @Success 200 {object} response.SuccessResponse{data=models.Team} "Team created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request body"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /teams [post]
// @Security BearerAuth
func (h *TeamHandler) CreateTeam(c echo.Context) error {
	// Get the user ID from the context (set by auth middleware)
	userID, err := authutil.GetUserIDFromContext(c)
	if err != nil || userID == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	// Decode the request body
	var createTeamRequest createTeamRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&createTeamRequest); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate the request body
	if err := validator.New().Struct(createTeamRequest); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body,"+err.Error())
	}

	// Begin the transaction
	tx, err := repository.StartTransaction(h.DB, c.Request().Context())
	if err != nil {
		return response.InternalServerError("Failed to begin transaction", err)
	}

	defer repository.DeferRollback(tx, c.Request().Context())

	// Generate team ID
	teamID, err := id.GetID()
	if err != nil {
		return response.InternalServerError("Failed to generate team ID", err)
	}

	// Generate team member ID
	teamMemberID, err := id.GetID()
	if err != nil {
		return response.InternalServerError("Failed to generate team member ID", err)
	}

	// Create the team
	team := models.Team{
		ID:        teamID,
		OwnerID:   *userID,
		Name:      createTeamRequest.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create the team member (owner)
	teamMember := models.TeamMember{
		ID:        teamMemberID,
		TeamID:    teamID,
		UserID:    *userID,
		Role:      models.RoleOwner,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create the team in the database
	if err = repository.CreateTeam(c.Request().Context(), tx, team); err != nil {
		return response.InternalServerError("Failed to create team", err)
	}

	// Add the owner as a team member
	if err = repository.CreateTeamMember(c.Request().Context(), tx, teamMember); err != nil {
		return response.InternalServerError("Failed to add team member", err)
	}

	// Commit the transaction
	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		return response.InternalServerError("Failed to commit transaction", err)
	}

	// Respond with the success message and team data
	return c.JSON(http.StatusOK, response.Success("Team created successfully", team))
}
