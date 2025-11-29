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
)

// +----------------------------------------------+
// | CreateDocument                               |
// +----------------------------------------------+

type createDocumentRequest struct {
	Name       string                `json:"name" validate:"required,min=1,max=255" example:"My Document"`
	Permission models.DocsPermission `json:"permission" validate:"required,oneof=private public public_write" example:"private"`
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

	docID, err := id.GetID()
	if err != nil {
		return response.InternalServerError("Failed to generate document ID", err)
	}

	now := time.Now()
	doc := models.Document{
		ID:         docID,
		OwnerID:    *userID,
		Name:       req.Name,
		Permission: req.Permission,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	tx, err := repository.StartTransaction(h.DB, c.Request().Context())
	if err != nil {
		return response.InternalServerError("Failed to begin transaction", err)
	}
	defer repository.DeferRollback(tx, c.Request().Context())

	if err := repository.CreateDocument(c.Request().Context(), tx, doc); err != nil {
		return response.InternalServerError("Failed to create document", err)
	}

	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		return response.InternalServerError("Failed to commit transaction", err)
	}

	return c.JSON(http.StatusOK, response.Success("Document created successfully", doc))
}
