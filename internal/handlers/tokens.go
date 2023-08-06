package handlers

import (
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"net/http"

	"github.com/labstack/echo/v4"
)

// RefreshTokenRequestBody is the format for the body of a request to refresh JWT tokens.
type RefreshTokenRequestBody struct {
	RefreshToken string `json:"refreshToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
}

// TokensResponseBody is the format for the response body containing refreshed JWT tokens.
type TokensResponseBody struct {
	AccessToken  string `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gDG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
	RefreshToken string `json:"refreshToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gDG9lIiwiaWF0IjoxNTE2MjM9MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
}

// RefreshTokens godoc
// @Summary refresh pair JWT tokens
// @Tags tokens
// @Accept  json
// @Produce application/json
// @Param login body RefreshTokenRequestBody true "raw request body, should contain Refresh Token"
// @Success 200 {object} Response{data=TokensResponseBody}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/token/refresh [post]
func (h *BaseHandler) RefreshTokens(ctx context.Context, c echo.Context) error {
	var requestPayload RefreshTokenRequestBody

	if err := c.Bind(&requestPayload); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body", err)
	}

	if requestPayload == (RefreshTokenRequestBody{}) {
		return ErrorResponse(c, http.StatusBadRequest, "invalid request body", errors.New("empty request body"))
	}

	refreshTokenClaims, err := h.tokenService.DecodeRefreshToken(requestPayload.RefreshToken)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid refresh token", err)
	}

	if value, _ := h.tokenService.GetCacheValue(ctx, refreshTokenClaims.RefreshUUID); value == nil {
		return ErrorResponse(c, http.StatusUnauthorized, "refresh token expired", fmt.Errorf("refresh token %s not found in cache", refreshTokenClaims.RefreshUUID))
	}

	_ = h.tokenService.DropCacheKey(ctx, refreshTokenClaims.RefreshUUID)

	ts, err := h.tokenService.CreateToken(refreshTokenClaims.UserID)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error creating tokens", err)
	}

	if err = h.tokenService.CreateCacheKey(ctx, refreshTokenClaims.UserID, ts); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error creating cache key", err)
	}

	tokens := TokensResponseBody{
		AccessToken:  ts.AccessToken,
		RefreshToken: ts.RefreshToken,
	}

	return SuccessResponse(c, http.StatusOK, fmt.Sprintf("tokens successfuly refreshed for User %d", refreshTokenClaims.UserID), tokens)
}
