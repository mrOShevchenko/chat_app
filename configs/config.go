package configs

import (
	"chat_app/internal/handlers"
	"chat_app/internal/models"
	"chat_app/internal/services/tokenService"
)

type Config struct {
	BaseHandler  *handlers.BaseHandler
	UserRepo     models.UserRepository
	TokenService *tokenService.Service
}
