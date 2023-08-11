package internal

import (
	"chat_app/configs"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type AppConfig struct {
	Config *configs.Config
}

func (app *AppConfig) NewRouter() *echo.Echo {
	e := echo.New()

	e.GET("/docs/*", echoSwagger.WrapHandler)

	// User
	e.GET("/v1/user", app.Config.BaseHandler.User, app.AuthUserMiddleware())
	e.PUT("/v1/user/update-password", app.Config.BaseHandler.UpdatePassword, app.AuthUserMiddleware())
	e.PUT("/v1/user/update-username", app.Config.BaseHandler.UpdateUsername, app.AuthUserMiddleware())
	e.POST("/v1/user/upload-image", app.Config.BaseHandler.UploadImage, app.AuthUserMiddleware())
	e.POST("/v1/user/:user_id/follow", app.Config.BaseHandler.Follow, app.AuthUserMiddleware())
	e.DELETE("/v1/user/:user_id/follow", app.Config.BaseHandler.Unfollow, app.AuthUserMiddleware())
	e.POST("/v1/user/:user_id/block", app.Config.BaseHandler.Block, app.AuthUserMiddleware())
	e.DELETE("/v1/user/:user_id/block", app.Config.BaseHandler.Unblock, app.AuthUserMiddleware())

	// Auth
	e.POST("/v1/auth/register", app.Config.BaseHandler.Registration)
	e.POST("/v1/auth/login", app.Config.BaseHandler.Login)
	e.POST("/v1/auth/logout", app.Config.BaseHandler.Logout, app.AuthTokenMiddleware())

	// Tokens
	e.POST("/v1/token/refresh", app.Config.BaseHandler.RefreshTokens)

	// Search
	e.GET("/v1/search", app.Config.BaseHandler.Search, app.AuthUserMiddleware())

	//Chat
	e.GET("/ws/chat", app.Config.BaseHandler.Chat)
	e.POST("/v1/user/:user_id/chat", app.Config.BaseHandler.CreateChat, app.AuthUserMiddleware())
	e.GET("/v1/user/:user_id/chat", app.Config.BaseHandler.GetChat, app.AuthUserMiddleware())
	e.DELETE("/v1/chat/:chat_id", app.Config.BaseHandler.DeleteChat, app.AuthUserMiddleware())
	e.GET("/v1/chat/:chat_id/messages", app.Config.BaseHandler.GetMessages, app.AuthUserMiddleware())

	// Device
	e.POST("/v1/device", app.Config.BaseHandler.AddDevice, app.AuthUserMiddleware())

	return e
}
