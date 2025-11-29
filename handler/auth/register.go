package auth

import (
	"encoding/json"
	"net/http"
	"ridash/repository"
	"ridash/utils/response"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// +----------------------------------------------+
// | Register                                     |
// +----------------------------------------------+

type registerRequest struct {
	Email       string `json:"email" validate:"required,email,max=255" example:"user@example.com"`
	Password    string `json:"password" validate:"required,min=8,max=255" example:"password123"`
	DisplayName string `json:"display_name" validate:"required,min=3,max=255" example:"John Doe"`
}

// Register godoc
// @Summary Register a new user
// @Description Creates a new user account with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body registerRequest true "Registration request"
// @Success 200 {object} response.SuccessResponse "User registered successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request body or email already in use"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	// Decode the request body
	var registerRequest registerRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&registerRequest); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate the request body
	if err := validator.New().Struct(registerRequest); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body,"+err.Error())
	}

	// Begin the transaction
	tx, err := repository.StartTransaction(h.DB, c.Request().Context())
	if err != nil {
		return response.InternalServerError("Failed to begin transaction", err)
	}

	defer repository.DeferRollback(tx, c.Request().Context())

	// Get the account by email
	checkedAccount, err := repository.GetAccountByEmail(c.Request().Context(), tx, registerRequest.Email)
	if err != nil {
		return response.InternalServerError("Failed to check if user already exists", err)
	}

	// If the account is found, return an error
	if checkedAccount != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "This email is already in use")
	}

	// Generate the user and account
	user, account, err := GenerateUser(registerRequest)
	if err != nil {
		return response.InternalServerError("Failed to generate user", err)
	}

	// Create the user and account in the database
	if err = repository.CreateUserAndAccount(c.Request().Context(), tx, user, account); err != nil {
		return response.InternalServerError("Failed to create user", err)
	}

	// Generate the refresh token
	refreshToken, err := generateTokenAndSaveRefreshToken(c, tx, user.ID)
	if err != nil {
		return response.InternalServerError("Failed to generate refresh token", err)
	}

	// Commit the transaction
	if err := repository.CommitTransaction(tx, c.Request().Context()); err != nil {
		return response.InternalServerError("Failed to commit transaction", err)
	}

	// Generate the refresh token cookie
	refreshTokenCookie := generateRefreshTokenCookie(refreshToken)
	c.SetCookie(&refreshTokenCookie)

	// Respond with the success message
	return c.JSON(http.StatusOK, response.SuccessMessage("User registered successfully"))
}
