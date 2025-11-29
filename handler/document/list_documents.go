package document

import (
	"net/http"
	"ridash/repository"
	authutil "ridash/utils/auth"
	"ridash/utils/response"

	"github.com/labstack/echo/v4"
)

// +----------------------------------------------+
// | ListDocuments                                |
// +----------------------------------------------+

// ListDocuments godoc
// @Summary List documents
// @Description Lists documents visible to the current user (public, owned, or shared)
// @Tags documents
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=[]models.Document} "Documents retrieved successfully"
// @Failure 401 {object} response.ErrorResponse "Invalid user ID"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /documents [get]
func (h *DocumentHandler) ListDocuments(c echo.Context) error {
	userID, err := authutil.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user ID")
	}

	tx, err := repository.StartTransaction(h.DB, c.Request().Context())
	if err != nil {
		return response.InternalServerError("Failed to begin transaction", err)
	}
	defer repository.DeferRollback(tx, c.Request().Context())

	documents, err := repository.ListDocumentsForUser(c.Request().Context(), tx, userID)
	if err != nil {
		return response.InternalServerError("Failed to list documents", err)
	}

	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		return response.InternalServerError("Failed to commit transaction", err)
	}

	return c.JSON(http.StatusOK, response.Success("Documents retrieved successfully", documents))
}
