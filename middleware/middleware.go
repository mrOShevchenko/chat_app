package middleware

import (
	"chat_app/configs"
	"chat_app/internal/handlers"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"os"
)

type MiddlewareConfig struct {
	config *configs.Config
}

func (app *MiddlewareConfig) AddMiddleware(e *echo.Echo) {
	DeafaulCORSConfig := middleware.CORSConfig{
		Skipper:      middleware.DefaultSkipper,
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet,
			http.MethodHead, http.MethodPut, http.MethodPost,
			http.MethodDelete, http.MethodOptions},
	}
	e.Use(middleware.CORSWithConfig(DeafaulCORSConfig))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
}

func (app *MiddlewareConfig) AuthTokenMiddleware() echo.MiddlewareFunc {
	AuthorizationConfig := echojwt.Config{
		SigningKey:     []byte(os.Getenv("ACCESS_SECRET")),
		ParseTokenFunc: app.ParseToken,
		ErrorHandler: func(c echo.Context, err error) error {
			return handlers.ErrorResponse(c, http.StatusUnauthorized, "invalid token", err)
		},
	}

	return echojwt.WithConfig(AuthorizationConfig)
}

func (app *MiddlewareConfig) AuthUserMiddleware() echo.MiddlewareFunc {
	AuthorizationConfig := echojwt.Config{
		SigningKey:     []byte(os.Getenv("ACCESS_SECRET")),
		ParseTokenFunc: app.GetUser,
		ErrorHandler: func(c echo.Context, err error) error {
			return handlers.ErrorResponse(c, http.StatusUnauthorized, "invalid token", err)
		},
	}

	return echojwt.WithConfig(AuthorizationConfig)
}

func (app *MiddlewareConfig) ParseToken(c echo.Context, auth string) (interface{}, error) {
	_ = c

	accessTokenClaims, err := app.config.TokenService.DecodeAccessToken(auth)
	if err != nil {
		return nil, err
	}

	_, err = app.config.TokenService.GetCacheValue(c.Request().Context(), accessTokenClaims.AccessUUID)
	if err != nil {
		return nil, err
	}

	return accessTokenClaims, nil
}

func (app *MiddlewareConfig) GetUser(c echo.Context, auth string) (interface{}, error) {
	_ = c
	accessTokenClaims, err := app.config.TokenService.DecodeAccessToken(auth)
	if err != nil {
		return nil, err
	}

	_, err = app.config.TokenService.GetCacheValue(c.Request().Context(), accessTokenClaims.AccessUUID)
	if err != nil {
		return nil, err
	}

	user, err := app.config.UserRepo.FindByID(accessTokenClaims.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
