package handlers

import (
	"chat_app/internal/models"
	"chat_app/internal/services"
	"chat_app/internal/services/tokenService"
	"github.com/labstack/echo/v4"
	"os"
)

type BaseHandler struct {
	userRepo     models.UserRepository
	messageRepo  models.MessageRepository
	chatRepo     models.ChatRepository
	tokenService *tokenService.Service
	chatService  *services.ChatService
}

// Response is the type used for sending JSON around
type Response struct {
	Error   bool        `json:"error" example:"false"`
	Message string      `json:"message" example:"success operation"`
	Data    interface{} `json:"data,omitempty"`
}

// NewBaseHandler is a constructor for BaseHandler
func NewBaseHandler(
	userRepo models.UserRepository,
	messageRepo models.MessageRepository,
	chatRepo models.ChatRepository,
	tokenService *tokenService.Service,
	chatService *services.ChatService,
) *BaseHandler {
	return &BaseHandler{userRepo: userRepo,
		messageRepo:  messageRepo,
		chatRepo:     chatRepo,
		tokenService: tokenService,
		chatService:  chatService}
}

// SuccessResponse creates a JSON response with success status.
func SuccessResponse(c echo.Context, statusCode int, message string, data any) error {
	payload := Response{
		Error:   false,
		Message: message,
		Data:    data,
	}
	return c.JSON(statusCode, payload)
}

// ErrorResponse creates a JSON response with error status.
func ErrorResponse(c echo.Context, statusCode int, message string, err error) error {
	payload := Response{
		Error:   true,
		Message: message,
	}

	// Including error details if running in 'dev' mode.
	if os.Getenv("APP_ENV") == "dev" {
		payload.Data = err.Error()
	}

	return c.JSON(statusCode, payload)
}
