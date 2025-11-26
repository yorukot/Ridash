package auth

import (
	"context"
	"net/http"
	"ridash/models"
	"ridash/repository"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

// +----------------------------------------------+
// | OAuth Callback                               |
// +----------------------------------------------+

// OAuthCallback godoc
// @Summary OAuth callback handler
// @Description Handles OAuth provider callback, processes authorization code, creates/links user accounts, and issues authentication tokens
// @Tags oauth
// @Accept json
// @Produce json
// @Param provider path string true "OAuth provider (e.g., google, github)"
// @Param code query string true "Authorization code from OAuth provider"
// @Param state query string true "OAuth state parameter for CSRF protection"
// @Success 307 {string} string "Redirect to success URL with authentication cookies set"
// @Failure 400 {object} response.ErrorResponse "Invalid provider, oauth state, or verification failed"
// @Failure 500 {object} response.ErrorResponse "Internal server error during user creation or token generation"
// @Router /auth/oauth/{provider}/callback [get]
func (h *AuthHandler) OAuthCallback(c echo.Context) error {
	// Get the oauth state from the query params
	oauthState := c.QueryParam("state")
	code := c.QueryParam("code")

	// Get the oauth session cookie
	oauthSessionCookie, err := c.Cookie(models.CookieNameOAuthSession)
	if err != nil {
		c.Logger().Debugf("OAuth session cookie not found: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "OAuth session not found")
	}

	// Parse the provider
	provider, err := parseProvider(c.Param("provider"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid provider")
	}

	// No need to check if the provider is valid because it's checked in the parseProvider function
	oauthConfig := h.OAuthConfig.Providers[provider]
	// Get the oidc provider
	oidcProvider := h.OAuthConfig.OIDCProviders[provider]

	// Validate the oauth state
	valid, payload, err := oauthValidateStateWithPayload(oauthSessionCookie.Value)
	if err != nil || !valid || oauthState != payload.State {
		c.Logger().Warnf("OAuth state validation failed - ip: %s, user_agent: %s, provider: %s, oauth_state: %s, payload_state: %s",
			c.RealIP(),
			c.Request().UserAgent(),
			string(provider),
			oauthState,
			payload.State)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid oauth state")
	}

	// Get the user ID from the session cookie
	var userID int64
	var accountID int64
	if payload.Subject != "" {
		userID, err = strconv.ParseInt(payload.Subject, 10, 64)
		if err != nil {
			c.Logger().Errorf("Failed to parse user ID: %v", err)
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID in session")
		}
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	// Exchange the code for a token
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to exchange code")
	}

	// Get the raw ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		c.Logger().Error("Failed to get id token")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get id token")
	}

	// Verify the token
	userInfo, err := oauthVerifyTokenAndGetUserInfo(c.Request().Context(), rawIDToken, token, oidcProvider, oauthConfig)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to verify token")
	}

	// Begin the transaction
	tx, err := repository.StartTransaction(h.DB, c.Request().Context())
	if err != nil {
		c.Logger().Errorf("Failed to begin transaction: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to begin transaction")
	}

	defer repository.DeferRollback(tx, c)

	// Get the account and user by the provider and user ID for checking if the user is already linked/registered
	account, user, err := repository.GetAccountWithUserByProviderUserID(c.Request().Context(), tx, provider, userInfo.Subject)
	if err != nil {
		c.Logger().Errorf("Failed to get account: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get account")
	}

	// If the account is not found and the userID is not zero, it means the user is already registered
	// so we need to link the account to the user
	if user == nil && userID != 0 {
		// Link the account to the user
		newAccount, err := generateUserAccountFromOAuthUserInfo(userInfo, provider, userID)
		if err != nil {
			c.Logger().Errorf("Failed to link account: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate account")
		}

		accountID = newAccount.ID

		// Create the account
		if err = repository.CreateAccount(c.Request().Context(), tx, newAccount); err != nil {
			c.Logger().Errorf("Failed to create account: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create account")
		}

		c.Logger().Infof("OAuth link account successful - provider: %s, user_id: %d, ip: %s", string(provider), userID, c.RealIP())
	} else if account == nil && userID == 0 {
		// Generate the full user object
		newUser, newAccount, err := generateUserFromOAuthUserInfo(userInfo, provider)
		if err != nil {
			c.Logger().Errorf("Failed to generate user: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate user and account")
		}

		// Create the user and account
		if err = repository.CreateUserAndAccount(c.Request().Context(), tx, newUser, newAccount); err != nil {
			c.Logger().Errorf("Failed to create user: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user and account")
		}

		accountID = newAccount.ID
		userID = newUser.ID

		c.Logger().Infof("OAuth new user registered - provider: %s, user_id: %d, ip: %s", string(provider), userID, c.RealIP())
	} else {
		accountID = account.ID
		userID = user.ID

		c.Logger().Infof("OAuth login successful - provider: %s, user_id: %d, ip: %s", string(provider), userID, c.RealIP())
	}

	// If the user ID is zero, it means something went wrong (it should not happen)
	if userID == 0 {
		c.Logger().Errorf("User ID is zero - user: %v", user)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user and account")
	}

	// Create the oauth token
	err = repository.CreateOAuthToken(c.Request().Context(), tx, models.OAuthToken{
		AccountID:    accountID,
		AccessToken:  token.AccessToken,
		RefreshToken: &token.RefreshToken,
		Expiry:       token.Expiry,
		TokenType:    token.TokenType,
		Provider:     provider,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	})
	if err != nil {
		c.Logger().Errorf("Failed to create oauth token: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create oauth token")
	}

	// Generate the refresh token
	refreshToken, err := generateTokenAndSaveRefreshToken(c, tx, userID)
	if err != nil {
		c.Logger().Errorf("Failed to create refresh token: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create refresh token")
	}

	// Commit the transaction
	if err := repository.CommitTransaction(tx, c); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit transaction")
	}

	// Generate the refresh token cookie
	refreshTokenCookie := generateRefreshTokenCookie(refreshToken)
	c.SetCookie(&refreshTokenCookie)

	// Redirect to the redirect URI
	return c.Redirect(http.StatusTemporaryRedirect, payload.RedirectURI)
}
