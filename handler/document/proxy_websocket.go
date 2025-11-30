package document

import (
	"context"
	"errors"
	"net/http"
	"net/http/httputil"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"ridash/models"
	"ridash/repository"
	authutil "ridash/utils/auth"
	"ridash/utils/docmanager"
)

// ProxyDocumentWebsocket upgrades the connection and proxies it to the document manager.
func (h *DocumentHandler) ProxyDocumentWebsocket(c echo.Context) error {
	if h.DocManager == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Document manager unavailable")
	}

	docID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid document ID")
	}

	userID, err := authutil.GetUserIDFromContext(c)
	if err != nil || userID == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	ctx := c.Request().Context()
	tx, err := repository.StartTransaction(h.DB, ctx)
	if err != nil {
		zap.L().Error("Failed to begin transaction", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to begin transaction")
	}
	defer repository.DeferRollback(tx, ctx)

	docCtx, err := loadDocumentContext(ctx, tx, docID)
	if err != nil {
		zap.L().Error("Failed to get document", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get document")
	}

	if docCtx.Document == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Document not found")
	}

	doc := docCtx.Document
	allowed, err := canWriteDocument(ctx, tx, docCtx, *userID)
	if err != nil {
		zap.L().Error("Failed to check permissions", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to check permissions")
	}

	if !allowed {
		return echo.NewHTTPError(http.StatusForbidden, "Access denied")
	}

	if err := repository.CommitTransaction(tx, ctx); err != nil {
		zap.L().Error("Failed to commit transaction", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit transaction")
	}

	ticket, err := h.DocManager.IssueTicket(ctx, doc.ID, strconv.FormatInt(*userID, 10))
	if err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, docmanager.ErrDocumentNotFound) {
			status = http.StatusNotFound
		}

		zap.L().Error("Failed to issue document ticket", zap.Error(err), zap.Int64("document_id", doc.ID), zap.Int64("user_id", *userID))
		return echo.NewHTTPError(status, "Failed to open document session")
	}

	target, err := h.DocManager.EditEndpoint(ticket.Ticket)
	if err != nil {
		zap.L().Error("Failed to prepare document session", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to prepare document session")
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = target.Path
			req.URL.RawPath = target.RawPath
			req.URL.RawQuery = target.RawQuery
			req.Host = target.Host
		},
		ErrorHandler: func(rw http.ResponseWriter, req *http.Request, proxyErr error) {
			if errors.Is(proxyErr, context.Canceled) {
				return
			}
			zap.L().Error("Document websocket proxy failed", zap.Error(proxyErr), zap.Int64("document_id", doc.ID), zap.Int64("user_id", *userID))
			http.Error(rw, "Failed to proxy websocket", http.StatusBadGateway)
		},
	}

	proxy.ServeHTTP(c.Response(), c.Request())
	return nil
}

func canWriteDocument(ctx context.Context, tx pgx.Tx, docCtx documentContext, userID int64) (bool, error) {
	if isTeamOwner(userID, docCtx.Team) {
		return true, nil
	}

	share, err := repository.GetShareByDocumentAndUser(ctx, tx, docCtx.Document.ID, userID)
	if err != nil {
		return false, err
	}

	if share != nil && share.Roles == models.DocsSharePermissionWrite {
		return true, nil
	}

	if docCtx.Document.Permission == models.DocsPermissionPublicWrite {
		return true, nil
	}

	return false, nil
}
