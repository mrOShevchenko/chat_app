package internal

import (
	"chat_app/configs"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type Config struct {
	config *configs.Config
}

func (app *Config) NewRouter() *echo.Echo {
	e := echo.New()

	e.GET("/docs/*", echoSwagger.WrapHandler)

	// User
	e.GET("/v1/user", app.config.BaseHandler.User, app.AuthUserMiddleware())
	e.PUT("/v1/user/update-password", app.BaseHandler.UpdatePassword, app.AuthUserMiddleware())
	e.PUT("/v1/user/update-username", app.BaseHandler.UpdateUsername, app.AuthUserMiddleware())
	e.POST("/v1/user/upload-image", app.BaseHandler.UploadImage, app.AuthUserMiddleware())
	e.POST("/v1/user/:user_id/follow", app.BaseHandler.Follow, app.AuthUserMiddleware())
	e.DELETE("/v1/user/:user_id/follow", app.BaseHandler.Unfollow, app.AuthUserMiddleware())
	e.POST("/v1/user/:user_id/block", app.BaseHandler.Block, app.AuthUserMiddleware())
	e.DELETE("/v1/user/:user_id/block", app.BaseHandler.Unblock, app.AuthUserMiddleware())

	// Auth
	e.POST("/v1/auth/register", app.BaseHandler.Registration)
	e.POST("/v1/auth/login", app.BaseHandler.Login)
	e.POST("/v1/auth/logout", app.BaseHandler.Logout, app.AuthTokenMiddleware())

	// Tokens
	e.POST("/v1/token/refresh", app.BaseHandler.RefreshTokens)

	// Search
	e.GET("/v1/search", app.BaseHandler.Search, app.AuthUserMiddleware())

	//Chat
	e.GET("/ws/chat", app.BaseHandler.Chat)
	e.POST("/v1/user/:user_id/chat", app.BaseHandler.CreateChat, app.AuthUserMiddleware())
	e.GET("/v1/user/:user_id/chat", app.BaseHandler.GetChat, app.AuthUserMiddleware())
	e.DELETE("/v1/chat/:chat_id", app.BaseHandler.DeleteChat, app.AuthUserMiddleware())
	e.GET("/v1/chat/:chat_id/messages", app.BaseHandler.GetMessages, app.AuthUserMiddleware())

	// Device
	e.POST("/v1/device", app.BaseHandler.AddDevice, app.AuthUserMiddleware())

	return e
}
