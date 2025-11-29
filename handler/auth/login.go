package auth

import (
	"encoding/json"
	"net/http"
	"ridash/repository"
	"ridash/utils/encrypt"
	"ridash/utils/response"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// +----------------------------------------------+
// | Login                                        |
// +----------------------------------------------+

// LoginRequest is the request body for the login endpoint
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=255"`
}

// Login godoc
// @Summary User login
// @Description Authenticates a user with email and password, returns a refresh token cookie
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request with email and password"
// @Success 200 {object} response.SuccessResponse "Login successful, refresh token set in cookie"
// @Failure 400 {object} response.ErrorResponse "Invalid request body or invalid credentials"
// @Failure 500 {object} response.ErrorResponse "Internal server error (transaction, database, or password verification failure)"
// @Failure 502 {object} response.ErrorResponse "Invalid request body format"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	// Decode the request body
	var loginRequest LoginRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&loginRequest); err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "Invalid request body")
	}

	// Validate the request body
	if err := validator.New().Struct(loginRequest); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body,"+err.Error())
	}

	// Begin the transaction
	tx, err := repository.StartTransaction(h.DB, c.Request().Context())
	if err != nil {
		return response.InternalServerError("Failed to begin transaction", err)
	}
	defer repository.DeferRollback(tx, c.Request().Context())

	// Get the user by email
	user, err := repository.GetUserByEmail(c.Request().Context(), tx, loginRequest.Email)
	if err != nil {
		return response.InternalServerError("Failed to get user by email", err)
	}

	// TODO: Need to change this
	// If the user is not found, return an error
	if user == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid credentials")
	}

	// Compare the password and hash
	match, err := encrypt.ComparePasswordAndHash(loginRequest.Password, *user.PasswordHash)
	if err != nil {
		return response.InternalServerError("Failed to compare password and hash", err)
	}

	// If the password is not correct, return an error
	if !match {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid credentials")
	}

	// Generate the refresh token
	refreshToken, err := generateTokenAndSaveRefreshToken(c, tx, user.ID)
	if err != nil {
		return response.InternalServerError("Failed to generate refresh token", err)
	}

	// Commit the transaction
	repository.CommitTransaction(tx, c.Request().Context())

	// Generate the refresh token cookie
	refreshTokenCookie := generateRefreshTokenCookie(refreshToken)
	c.SetCookie(&refreshTokenCookie)

	return c.JSON(http.StatusOK, response.SuccessMessage("Login successful"))
}
