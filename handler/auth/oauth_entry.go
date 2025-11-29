package auth

import (
	"net/http"
	authutil "ridash/utils/auth"
	"ridash/utils/config"
	"ridash/utils/response"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

// +----------------------------------------------+
// | OAuth Entry                                  |
// +----------------------------------------------+

// OAuthEntry godoc
// @Summary Initiate OAuth flow
// @Description Redirects user to OAuth provider for authentication
// @Tags oauth
// @Param provider path string true "OAuth provider (e.g., google, github)"
// @Param next query string false "Redirect URL after successful OAuth linking"
// @Success 307 {string} string "Redirect to OAuth provider"
// @Failure 400 {object} response.ErrorResponse "Invalid provider or bad request"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /auth/oauth/{provider} [get]
func (h *AuthHandler) OAuthEntry(c echo.Context) error {
	// Parse provider
	provider, err := parseProvider(c.Param("provider"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid provider")
	}

	var userID string
	if id, err := authutil.GetUserIDFromContext(c); err == nil && id != nil {
		userID = strconv.FormatInt(*id, 10)
	}

	expiresAt := time.Now().Add(time.Duration(config.Env().OAuthStateExpiresAt) * time.Second)

	next := c.QueryParam("next")
	if next == "" {
		next = "/"
	}

	oauthStateJwt, oauthState, err := oauthGenerateStateWithPayload(next, expiresAt, userID)
	if err != nil {
		return response.InternalServerError("Failed to generate oauth state", err)
	}

	oauthConfig := h.OAuthConfig.Providers[provider]

	authURL := oauthConfig.AuthCodeURL(
		oauthState,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	)

	oauthSessionCookie := generateOAuthSessionCookie(oauthStateJwt)
	c.SetCookie(&oauthSessionCookie)

	return c.Redirect(http.StatusTemporaryRedirect, authURL)
}
