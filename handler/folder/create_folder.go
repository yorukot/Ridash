package folder

import (
	"encoding/json"
	"net/http"
	"ridash/models"
	"ridash/repository"
	authutil "ridash/utils/auth"
	"ridash/utils/id"
	"ridash/utils/response"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// +----------------------------------------------+
// | CreateFolder                                 |
// +----------------------------------------------+

type createFolderRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=255" example:"Project Docs"`
	ParentFolder *int64 `json:"parent_folder,omitempty" validate:"omitempty,gt=0" example:"175928847299117063"`
}

// CreateFolder godoc
// @Summary Create a new folder for a team
// @Description Creates a folder within a team; only team owner can create folders
// @Tags folder
// @Accept json
// @Produce json
// @Param teamID path int true "Team ID"
// @Param request body createFolderRequest true "Create folder request"
// @Success 200 {object} response.SuccessResponse{data=models.Folder} "Folder created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request body or team ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Only team owner can create folders"
// @Failure 404 {object} response.ErrorResponse "Team or parent folder not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /teams/{teamID}/folders [post]
// @Security BearerAuth
func (h *FolderHandler) CreateFolder(c echo.Context) error {
	userID, err := authutil.GetUserIDFromContext(c)
	if err != nil || userID == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	teamIDStr := c.Param("teamID")
	teamID, err := strconv.ParseInt(teamIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid team ID")
	}

	var req createFolderRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := validator.New().Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body,"+err.Error())
	}

	tx, err := repository.StartTransaction(h.DB, c.Request().Context())
	if err != nil {
		zap.L().Error("Failed to begin transaction", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to begin transaction")
	}
	defer repository.DeferRollback(tx, c.Request().Context())

	team, err := repository.GetTeamByID(c.Request().Context(), tx, teamID)
	if err != nil {
		zap.L().Error("Failed to get team", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get team")
	}

	if team == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Team not found")
	}

	if team.OwnerID != *userID {
		return echo.NewHTTPError(http.StatusForbidden, "Only team owner can create folders")
	}

	if req.ParentFolder != nil {
		parentFolder, err := repository.GetFolderByIDAndTeamID(c.Request().Context(), tx, *req.ParentFolder, teamID)
		if err != nil {
			zap.L().Error("Failed to get parent folder", zap.Error(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get parent folder")
		}

		if parentFolder == nil {
			return echo.NewHTTPError(http.StatusNotFound, "Parent folder not found")
		}
	}

	folderID, err := id.GetID()
	if err != nil {
		zap.L().Error("Failed to generate folder ID", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate folder ID")
	}

	now := time.Now()

	folder := models.Folder{
		ID:           folderID,
		TeamID:       teamID,
		Name:         req.Name,
		ParentFolder: req.ParentFolder,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := repository.CreateFolder(c.Request().Context(), tx, folder); err != nil {
		zap.L().Error("Failed to create folder", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create folder")
	}

	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		zap.L().Error("Failed to commit transaction", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit transaction")
	}

	return c.JSON(http.StatusOK, response.Success("Folder created successfully", folder))
}
