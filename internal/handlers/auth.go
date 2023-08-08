package handlers

import (
	"chat_app/internal/models"
	"chat_app/internal/services/tokenService"
	"chat_app/pkg/validators"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"os"
)

// RegistrationRequestBody represents request data for the registration process.
type RegistrationRequestBody struct {
	Username string `json:"username" example:"username"`
	Password string `json:"password" example:"testPassword"`
}

// LoginRequestBody represents request data for the login process.
type LoginRequestBody struct {
	Username string `json:"username" example:"username"`
	Password string `json:"password" example:"testPassword"`
}

// Registration godoc
// @Summary registration user by credentials
// @Description Username should contain:
// @Description - lower, upper case latin letters and digits
// @Description - minimum 8 characters length
// @Description - maximum 40 characters length
// @Description Password should contain:
// @Description - minimum of one small case letter
// @Description - minimum of one upper case letter
// @Description - minimum of one digit
// @Description - minimum of one special character
// @Description - minimum 8 characters length
// @Description - maximum 40 characters length
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param registration body RegistrationRequestBody true "raw request body"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /v1/auth/register [post]
func (h *BaseHandler) Registration(c echo.Context) error {
	var requestPayload RegistrationRequestBody

	// Parse and validate request payload
	if err := c.Bind(&requestPayload); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
	}

	if err := validators.ValidateUsername(requestPayload.Username); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid username", err)
	}

	if err := validators.ValidatePassword(requestPayload.Password); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid password", err)
	}

	// Check for duplicate username
	user, _ := h.userRepo.FindByUsername(requestPayload.Username)
	if user != nil {
		err := fmt.Errorf("user with username %s already exists", requestPayload.Username)
		return ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
	}

	// Hash the user's password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestPayload.Password), bcrypt.DefaultCost)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error hashing password", err)
	}

	// Create a new user
	newUser := models.User{
		Username: requestPayload.Username,
		Password: string(hashedPassword),
		Image:    os.Getenv("URL_PREFIX_IMAGES") + "default.png",
	}

	if err = h.userRepo.Create(&newUser); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error creating user", err)
	}

	return SuccessResponse(c, http.StatusOK, fmt.Sprintf("user with username: %s successfully created", requestPayload.Username), nil)
}

// Login godoc
// @Summary login user by credentials
// @Description Username should contain:
// @Description - lower, upper case latin letters and digits
// @Description - minimum 8 characters length
// @Description - maximum 40 characters length
// @Description Password should contain:
// @Description - minimum of one small case letter
// @Description - minimum of one upper case letter
// @Description - minimum of one digit
// @Description - minimum of one special character
// @Description - minimum 8 characters length
// @Description - maximum 40 characters length
// @Description Response contain pair JWT tokens, use /v1/tokens/refresh for updating them
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param login body LoginRequestBody true "raw request body"
// @Success 200 {object} Response{data=TokensResponseBody}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /v1/auth/login [post]
func (h *BaseHandler) Login(c echo.Context) error {
	var requestPayload LoginRequestBody

	// Parse and validate request payload
	if err := c.Bind(&requestPayload); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
	}

	if err := validators.ValidateUsername(requestPayload.Username); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid username", err)
	}

	if err := validators.ValidatePassword(requestPayload.Password); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid password", err)
	}

	// Verify the user exists and the password matches
	user, err := h.userRepo.FindByUsername(requestPayload.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrorResponse(c, http.StatusNotFound, "user not found", nil)
		}
		return ErrorResponse(c, http.StatusInternalServerError, "error while finding user", err)
	}

	if valid, err := h.userRepo.PasswordMatches(user, requestPayload.Password); err != nil || !valid {
		return ErrorResponse(c, http.StatusUnauthorized, "invalid credentials", err)
	}

	// Generate JWT tokens for the user
	ts, err := h.tokenService.CreateToken(user.ID)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error creating tokens", err)
	}

	// Add refresh token UUID to cache
	if err = h.tokenService.CreateCacheKey(c.Request().Context(), user.ID, ts); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error creating cache tokens", err)
	}

	tokens := TokensResponseBody{
		AccessToken:  ts.AccessToken,
		RefreshToken: ts.RefreshToken,
	}

	return SuccessResponse(c, http.StatusOK, fmt.Sprintf("user %s is logged in", requestPayload.Username), tokens)
}

// Logout godoc
// @Summary logout user
// @Tags auth
// @Accept  json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/auth/logout [post]
func (h *BaseHandler) Logout(c echo.Context) error {
	// receive AccessToken Claims from context middleware
	accessTokenClaims, ok := c.Get("user").(*tokenService.AccessTokenClaims)
	if !ok {
		err := errors.New("internal transport token error")
		return ErrorResponse(c, http.StatusInternalServerError, err.Error(), err)
	}

	if err := h.tokenService.DropCacheTokens(c.Request().Context(), *accessTokenClaims); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error clearing cache token", err)
	}

	return SuccessResponse(c, http.StatusOK, fmt.Sprintf("user %s successfully logged out", accessTokenClaims.AccessUUID), accessTokenClaims)
}
