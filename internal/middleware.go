package internal

import (
	"chat_app/internal/handlers"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"os"
)

func (app *AppConfig) AddMiddleware(e *echo.Echo) {
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

func (app *AppConfig) AuthTokenMiddleware() echo.MiddlewareFunc {
	AuthorizationConfig := echojwt.Config{
		SigningKey:     []byte(os.Getenv("ACCESS_SECRET")),
		ParseTokenFunc: app.ParseToken,
		ErrorHandler: func(c echo.Context, err error) error {
			return handlers.ErrorResponse(c, http.StatusUnauthorized, "invalid token", err)
		},
	}

	return echojwt.WithConfig(AuthorizationConfig)
}

func (app *AppConfig) AuthUserMiddleware() echo.MiddlewareFunc {
	AuthorizationConfig := echojwt.Config{
		SigningKey:     []byte(os.Getenv("ACCESS_SECRET")),
		ParseTokenFunc: app.GetUser,
		ErrorHandler: func(c echo.Context, err error) error {
			return handlers.ErrorResponse(c, http.StatusUnauthorized, "invalid token", err)
		},
	}

	return echojwt.WithConfig(AuthorizationConfig)
}

func (app *AppConfig) ParseToken(c echo.Context, auth string) (interface{}, error) {
	_ = c

	accessTokenClaims, err := app.Config.TokenService.DecodeAccessToken(auth)
	if err != nil {
		return nil, err
	}

	_, err = app.Config.TokenService.GetCacheValue(c.Request().Context(), accessTokenClaims.AccessUUID)
	if err != nil {
		return nil, err
	}

	return accessTokenClaims, nil
}

func (app *AppConfig) GetUser(c echo.Context, auth string) (interface{}, error) {
	_ = c
	accessTokenClaims, err := app.Config.TokenService.DecodeAccessToken(auth)
	if err != nil {
		return nil, err
	}

	_, err = app.Config.TokenService.GetCacheValue(c.Request().Context(), accessTokenClaims.AccessUUID)
	if err != nil {
		return nil, err
	}

	user, err := app.Config.UserRepo.FindByID(accessTokenClaims.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
