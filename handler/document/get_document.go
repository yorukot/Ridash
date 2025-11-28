package document

import (
	"errors"
	"net/http"
	"ridash/models"
	"ridash/repository"
	authutil "ridash/utils/auth"
	"ridash/utils/docmanager"
	"ridash/utils/response"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// +----------------------------------------------+
// | GetDocument                                  |
// +----------------------------------------------+

// GetDocument godoc
// @Summary Get document by ID
// @Description Retrieves a document if public or accessible by the authenticated user
// @Tags documents
// @Produce json
// @Param id path int true "Document ID"
// @Success 200 {object} response.SuccessResponse{data=models.DocumentWithContent} "Document retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid document ID"
// @Failure 401 {object} response.ErrorResponse "Invalid user ID"
// @Failure 403 {object} response.ErrorResponse "Access denied"
// @Failure 404 {object} response.ErrorResponse "Document not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /documents/{id} [get]
func (h *DocumentHandler) GetDocument(c echo.Context) error {
	docIDStr := c.Param("id")
	docID, err := strconv.ParseInt(docIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid document ID")
	}

	userID, err := authutil.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID")
	}

	tx, err := repository.StartTransaction(h.DB, c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to begin transaction")
	}
	defer repository.DeferRollback(tx, c.Request().Context())

	doc, err := repository.GetDocumentByID(c.Request().Context(), tx, docID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get document")
	}

	if doc == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Document not found")
	}

	// Enforce access for private documents
	if doc.Permission == models.DocsPermissionPrivate {
		if userID == nil {
			return echo.NewHTTPError(http.StatusForbidden, "Access denied")
		}

		if *userID != doc.OwnerID {
			share, err := repository.GetShareByDocumentAndUser(c.Request().Context(), tx, doc.ID, *userID)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to check share permissions")
			}
			if share == nil {
				return echo.NewHTTPError(http.StatusForbidden, "Access denied")
			}
		}
	}

	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit transaction")
	}

	result := any(doc)
	if h.DocManager != nil {
		content, err := h.DocManager.GetDocumentContent(c.Request().Context(), doc.ID)
		if err != nil && !errors.Is(err, docmanager.ErrDocumentNotFound) {
			zap.L().Error("Failed to fetch document content from manager", zap.Error(err), zap.Int64("document_id", doc.ID))
			return echo.NewHTTPError(http.StatusBadGateway, "Failed to fetch document content")
		}

		withContent := models.DocumentWithContent{
			Document: *doc,
		}

		if content != nil {
			withContent.Content = content.Content
			withContent.Seq = content.Seq
		}

		result = withContent
	}

	return c.JSON(http.StatusOK, response.Success("Document retrieved successfully", result))
}
