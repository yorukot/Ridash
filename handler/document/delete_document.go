package document

import (
	"errors"
	"net/http"
	"ridash/repository"
	authutil "ridash/utils/auth"
	"ridash/utils/docmanager"
	"ridash/utils/response"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// +----------------------------------------------+
// | DeleteDocument                               |
// +----------------------------------------------+

// DeleteDocument godoc
// @Summary Delete a document
// @Description Deletes a document owned by the authenticated user
// @Tags documents
// @Produce json
// @Param id path int true "Document ID"
// @Success 200 {object} response.SuccessResponse "Document deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid document ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "Document not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /documents/{id} [delete]
// @Security BearerAuth
func (h *DocumentHandler) DeleteDocument(c echo.Context) error {
	userID, err := authutil.GetUserIDFromContext(c)
	if err != nil || userID == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	docIDStr := c.Param("id")
	docID, err := strconv.ParseInt(docIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid document ID")
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

	if err := repository.DeleteSharesByDocument(c.Request().Context(), tx, docID); err != nil {
		return response.InternalServerError("Failed to delete document shares", err)
	}

	if err := repository.DeleteDocument(c.Request().Context(), tx, docID); err != nil {
		return response.InternalServerError("Failed to delete document", err)
	}

	if h.DocManager != nil {
		if err := h.DocManager.DeleteDocument(c.Request().Context(), docID); err != nil && !errors.Is(err, docmanager.ErrDocumentNotFound) {
			zap.L().Error("Failed to delete document from document manager", zap.Error(err), zap.Int64("document_id", docID))
			return echo.NewHTTPError(http.StatusBadGateway, "Failed to delete document content")
		}
	}

	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		return response.InternalServerError("Failed to commit transaction", err)
	}

	return c.JSON(http.StatusOK, response.SuccessMessage("Document deleted successfully"))
}
