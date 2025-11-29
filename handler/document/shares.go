package document

import (
	"encoding/json"
	"net/http"
	"ridash/models"
	"ridash/repository"
	authutil "ridash/utils/auth"
	"ridash/utils/id"
	"ridash/utils/response"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// +----------------------------------------------+
// | ListShares                                   |
// +----------------------------------------------+

// ListShares godoc
// @Summary List document shares
// @Description Lists all shares for a document (owner only)
// @Tags documents
// @Produce json
// @Param id path int true "Document ID"
// @Success 200 {object} response.SuccessResponse{data=[]models.DocsShare} "Shares retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid document ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "Document not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /documents/{id}/shares [get]
// @Security BearerAuth
func (h *DocumentHandler) ListShares(c echo.Context) error {
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

	shares, err := repository.ListSharesByDocument(c.Request().Context(), tx, docID)
	if err != nil {
		return response.InternalServerError("Failed to list shares", err)
	}

	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		return response.InternalServerError("Failed to commit transaction", err)
	}

	return c.JSON(http.StatusOK, response.Success("Shares retrieved successfully", shares))
}

// +----------------------------------------------+
// | CreateShare                                  |
// +----------------------------------------------+

type createShareRequest struct {
	UserID int64                      `json:"user_id,string" validate:"required" example:"175928847299117063"`
	Roles  models.DocsSharePermission `json:"roles" validate:"required,oneof=read write" example:"read"`
}

// CreateShare godoc
// @Summary Create a share
// @Description Shares a document with a user (owner only)
// @Tags documents
// @Accept json
// @Produce json
// @Param id path int true "Document ID"
// @Param request body createShareRequest true "Create share request"
// @Success 200 {object} response.SuccessResponse{data=models.DocsShare} "Share created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request body or document ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "Document not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /documents/{id}/shares [post]
// @Security BearerAuth
func (h *DocumentHandler) CreateShare(c echo.Context) error {
	userID, err := authutil.GetUserIDFromContext(c)
	if err != nil || userID == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	docIDStr := c.Param("id")
	docID, err := strconv.ParseInt(docIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid document ID")
	}

	var req createShareRequest
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

	existingShare, err := repository.GetShareByDocumentAndUser(c.Request().Context(), tx, docID, req.UserID)
	if err != nil {
		return response.InternalServerError("Failed to check existing share", err)
	}
	if existingShare != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Share already exists for this user")
	}

	shareID, err := id.GetID()
	if err != nil {
		return response.InternalServerError("Failed to generate share ID", err)
	}

	share := models.DocsShare{
		ID:         shareID,
		DocumentID: docID,
		UserID:     req.UserID,
		Roles:      req.Roles,
	}

	if err := repository.CreateShare(c.Request().Context(), tx, share); err != nil {
		return response.InternalServerError("Failed to create share", err)
	}

	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		return response.InternalServerError("Failed to commit transaction", err)
	}

	return c.JSON(http.StatusOK, response.Success("Share created successfully", share))
}

// +----------------------------------------------+
// | UpdateShare                                  |
// +----------------------------------------------+

type updateShareRequest struct {
	Roles models.DocsSharePermission `json:"roles" validate:"required,oneof=read write" example:"write"`
}

// UpdateShare godoc
// @Summary Update a share
// @Description Updates a document share (owner only)
// @Tags documents
// @Accept json
// @Produce json
// @Param id path int true "Document ID"
// @Param shareID path int true "Share ID"
// @Param request body updateShareRequest true "Update share request"
// @Success 200 {object} response.SuccessResponse{data=models.DocsShare} "Share updated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request body or path parameters"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "Document or share not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /documents/{id}/shares/{shareID} [put]
// @Security BearerAuth
func (h *DocumentHandler) UpdateShare(c echo.Context) error {
	userID, err := authutil.GetUserIDFromContext(c)
	if err != nil || userID == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	docID, shareID, err := parseDocumentAndShareIDs(c)
	if err != nil {
		return err
	}

	var req updateShareRequest
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

	share, err := repository.GetShareByID(c.Request().Context(), tx, shareID)
	if err != nil {
		return response.InternalServerError("Failed to get share", err)
	}
	if share == nil || share.DocumentID != docID {
		return echo.NewHTTPError(http.StatusNotFound, "Share not found")
	}

	if err := repository.UpdateShare(c.Request().Context(), tx, shareID, req.Roles); err != nil {
		return response.InternalServerError("Failed to update share", err)
	}

	share.Roles = req.Roles

	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		return response.InternalServerError("Failed to commit transaction", err)
	}

	return c.JSON(http.StatusOK, response.Success("Share updated successfully", share))
}

// +----------------------------------------------+
// | DeleteShare                                  |
// +----------------------------------------------+

// DeleteShare godoc
// @Summary Delete a share
// @Description Deletes a document share (owner only)
// @Tags documents
// @Produce json
// @Param id path int true "Document ID"
// @Param shareID path int true "Share ID"
// @Success 200 {object} response.SuccessResponse "Share deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid path parameters"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "Document or share not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /documents/{id}/shares/{shareID} [delete]
// @Security BearerAuth
func (h *DocumentHandler) DeleteShare(c echo.Context) error {
	userID, err := authutil.GetUserIDFromContext(c)
	if err != nil || userID == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	docID, shareID, err := parseDocumentAndShareIDs(c)
	if err != nil {
		return err
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

	share, err := repository.GetShareByID(c.Request().Context(), tx, shareID)
	if err != nil {
		return response.InternalServerError("Failed to get share", err)
	}
	if share == nil || share.DocumentID != docID {
		return echo.NewHTTPError(http.StatusNotFound, "Share not found")
	}

	if err := repository.DeleteShare(c.Request().Context(), tx, shareID); err != nil {
		return response.InternalServerError("Failed to delete share", err)
	}

	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		return response.InternalServerError("Failed to commit transaction", err)
	}

	return c.JSON(http.StatusOK, response.SuccessMessage("Share deleted successfully"))
}

// parseDocumentAndShareIDs parses document and share IDs from the path params.
func parseDocumentAndShareIDs(c echo.Context) (int64, int64, error) {
	docIDStr := c.Param("id")
	shareIDStr := c.Param("shareID")

	docID, err := strconv.ParseInt(docIDStr, 10, 64)
	if err != nil {
		return 0, 0, echo.NewHTTPError(http.StatusBadRequest, "Invalid document ID")
	}

	shareID, err := strconv.ParseInt(shareIDStr, 10, 64)
	if err != nil {
		return 0, 0, echo.NewHTTPError(http.StatusBadRequest, "Invalid share ID")
	}

	return docID, shareID, nil
}
