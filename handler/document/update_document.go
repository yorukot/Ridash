package document

import (
	"encoding/json"
	"net/http"
	"ridash/models"
	"ridash/repository"
	authutil "ridash/utils/auth"
	"ridash/utils/response"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// +----------------------------------------------+
// | UpdateDocument                               |
// +----------------------------------------------+

type updateDocumentRequest struct {
	Name       string                `json:"name" validate:"required,min=1,max=255" example:"Updated Document"`
	Permission models.DocsPermission `json:"permission" validate:"required,oneof=private public public_write" example:"public"`
}

// UpdateDocument godoc
// @Summary Update a document
// @Description Updates a document owned by the authenticated user
// @Tags documents
// @Accept json
// @Produce json
// @Param id path int true "Document ID"
// @Param request body updateDocumentRequest true "Update document request"
// @Success 200 {object} response.SuccessResponse{data=models.Document} "Document updated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request body or document ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "Document not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /documents/{id} [put]
// @Security BearerAuth
func (h *DocumentHandler) UpdateDocument(c echo.Context) error {
	userID, err := authutil.GetUserIDFromContext(c)
	if err != nil || userID == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	docIDStr := c.Param("id")
	docID, err := strconv.ParseInt(docIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid document ID")
	}

	var req updateDocumentRequest
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

	doc, err := repository.GetDocumentByID(c.Request().Context(), tx, docID)
	if err != nil {
		return response.InternalServerError("Failed to get document", err)
	}
	if doc == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Document not found")
	}
	if doc.OwnerID != *userID {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	now := time.Now()
	if err := repository.UpdateDocument(c.Request().Context(), tx, docID, req.Name, req.Permission, now); err != nil {
		return response.InternalServerError("Failed to update document", err)
	}

	doc.Name = req.Name
	doc.Permission = req.Permission
	doc.UpdatedAt = now

	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		return response.InternalServerError("Failed to commit transaction", err)
	}

	return c.JSON(http.StatusOK, response.Success("Document updated successfully", doc))
}
