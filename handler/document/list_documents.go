package document

import (
	"net/http"
	"ridash/repository"
	authutil "ridash/utils/auth"
	"ridash/utils/response"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
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
		zap.L().Error("Failed to begin transaction", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to begin transaction")
	}
	defer repository.DeferRollback(tx, c.Request().Context())

	documents, err := repository.ListDocumentsForUser(c.Request().Context(), tx, userID)
	if err != nil {
		zap.L().Error("Failed to list documents", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list documents")
	}

	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		zap.L().Error("Failed to commit transaction", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit transaction")
	}

	return c.JSON(http.StatusOK, response.Success("Documents retrieved successfully", documents))
}
