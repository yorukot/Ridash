package team

import (
	"net/http"
	"ridash/repository"
	authutil "ridash/utils/auth"
	"ridash/utils/response"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// +----------------------------------------------+
// | ListTeams                                    |
// +----------------------------------------------+

// ListTeams godoc
// @Summary List teams
// @Description Lists teams the authenticated user owns or belongs to
// @Tags team
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=[]models.Team} "Teams retrieved successfully"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /teams [get]
// @Security BearerAuth
func (h *TeamHandler) ListTeams(c echo.Context) error {
	userID, err := authutil.GetUserIDFromContext(c)
	if err != nil || userID == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	tx, err := repository.StartTransaction(h.DB, c.Request().Context())
	if err != nil {
		zap.L().Error("Failed to begin transaction", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to begin transaction")
	}
	defer repository.DeferRollback(tx, c.Request().Context())

	teams, err := repository.ListTeamsByUserID(c.Request().Context(), tx, *userID)
	if err != nil {
		zap.L().Error("Failed to list teams", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list teams")
	}

	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		zap.L().Error("Failed to commit transaction", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit transaction")
	}

	return c.JSON(http.StatusOK, response.Success("Teams retrieved successfully", teams))
}
