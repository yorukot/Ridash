package folder

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
// | UpdateFolder                                 |
// +----------------------------------------------+

type updateFolderRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=255" example:"Updated Folder Name"`
	ParentFolder *int64 `json:"parent_folder,omitempty" validate:"omitempty,gt=0" example:"175928847299117063"`
}

// UpdateFolder godoc
// @Summary Update a folder
// @Description Updates folder information (only accessible by team owner)
// @Tags folder
// @Accept json
// @Produce json
// @Param teamID path int true "Team ID"
// @Param id path int true "Folder ID"
// @Param request body updateFolderRequest true "Update folder request"
// @Success 200 {object} response.SuccessResponse{data=models.Folder} "Folder updated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request body or IDs"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Only team owner can update the folder"
// @Failure 404 {object} response.ErrorResponse "Team or folder not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /teams/{teamID}/folders/{id} [put]
// @Security BearerAuth
func (h *FolderHandler) UpdateFolder(c echo.Context) error {
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

	var req updateFolderRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := validator.New().Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body,"+err.Error())
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
		return echo.NewHTTPError(http.StatusForbidden, "Only team owner can update the folder")
	}

	folder, err := repository.GetFolderByIDAndTeamID(c.Request().Context(), tx, folderID, teamID)
	if err != nil {
		return response.InternalServerError("Failed to get folder", err)
	}

	if folder == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Folder not found")
	}

	if req.ParentFolder != nil {
		if *req.ParentFolder == folderID {
			return echo.NewHTTPError(http.StatusBadRequest, "A folder cannot be its own parent")
		}

		parentFolder, err := repository.GetFolderByIDAndTeamID(c.Request().Context(), tx, *req.ParentFolder, teamID)
		if err != nil {
			return response.InternalServerError("Failed to get parent folder", err)
		}

		if parentFolder == nil {
			return echo.NewHTTPError(http.StatusNotFound, "Parent folder not found")
		}
	}

	now := time.Now()

	if err := repository.UpdateFolder(c.Request().Context(), tx, folderID, teamID, req.Name, req.ParentFolder, now); err != nil {
		return response.InternalServerError("Failed to update folder", err)
	}

	folder.Name = req.Name
	folder.ParentFolder = req.ParentFolder
	folder.UpdatedAt = now

	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		return response.InternalServerError("Failed to commit transaction", err)
	}

	return c.JSON(http.StatusOK, response.Success("Folder updated successfully", folder))
}
