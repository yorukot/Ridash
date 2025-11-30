package document

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
	"go.uber.org/zap"
)

// +----------------------------------------------+
// | CreateDocument                               |
// +----------------------------------------------+

type createDocumentRequest struct {
	Name       string                `json:"name" validate:"required,min=1,max=255" example:"My Document"`
	Permission models.DocsPermission `json:"permission" validate:"required,oneof=private public public_write" example:"private"`
	FolderID   int64                 `json:"folder_id,string" validate:"required,gt=0" example:"175928847299117063"`
}

// CreateDocument godoc
// @Summary Create a document
// @Description Creates a new document owned by the authenticated user
// @Tags documents
// @Accept json
// @Produce json
// @Param request body createDocumentRequest true "Create document request"
// @Success 200 {object} response.SuccessResponse{data=models.Document} "Document created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request body"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /documents [post]
// @Security BearerAuth
func (h *DocumentHandler) CreateDocument(c echo.Context) error {
	userID, err := authutil.GetUserIDFromContext(c)
	if err != nil || userID == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	var req createDocumentRequest
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

	folder, err := repository.GetFolderByID(c.Request().Context(), tx, req.FolderID)
	if err != nil {
		zap.L().Error("Failed to get folder", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get folder")
	}
	if folder == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Folder not found")
	}

	team, err := repository.GetTeamByID(c.Request().Context(), tx, folder.TeamID)
	if err != nil {
		zap.L().Error("Failed to get team", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get team")
	}
	if team == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Team not found for folder")
	}
	if team.OwnerID != *userID {
		return echo.NewHTTPError(http.StatusForbidden, "Only team owner can create documents")
	}

	docID, err := id.GetID()
	if err != nil {
		zap.L().Error("Failed to generate document ID", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate document ID")
	}

	now := time.Now()
	doc := models.Document{
		ID:         docID,
		FolderID:   req.FolderID,
		Name:       req.Name,
		Permission: req.Permission,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := repository.CreateDocument(c.Request().Context(), tx, doc); err != nil {
		zap.L().Error("Failed to create document", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create document")
	}

	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		zap.L().Error("Failed to commit transaction", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit transaction")
	}

	return c.JSON(http.StatusOK, response.Success("Document created successfully", doc))
}
