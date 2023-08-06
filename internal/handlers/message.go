package handlers

import (
	"chat_app/internal/models"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

// GetMessages godoc
// @Summary Get messages from chat by ChatId
// @Tags chat
// @Accept  json
// @Produce application/json
// @Param	chat_id	path	int	true	"Chat ID"
// @Security BearerAuth
// @Success 200 {object} Response{data=[]models.Message}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /v1/chat/messages/{chat_id} [get]

func (h *BaseHandler) GetMessages(c echo.Context) error {
	user := c.Get("user").(*models.User)
	chatID, err := strconv.Atoi(c.Param("chat_id"))
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid chat ID", err)
	}

	chat, err := h.chatRepo.FindByID(chatID)
	if err != nil || chat == nil {
		return ErrorResponse(c, http.StatusNotFound, "Chat not found", err)
	}

	if !userInChat(user, chat) {
		err = fmt.Errorf("User with id %d not found in chat with id %d", user.ID, chatID)
		return ErrorResponse(c, http.StatusForbidden, "User not found in chat", err)
	}

	limit, err := getLimit(c)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid limit, must be an integer between 1 and 1000", err)
	}

	from, err := getFrom(c)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid 'from' parameter, must be an integer", err)
	}

	messages, err := h.messageRepo.GetMessages(chatID, from, limit)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to get messages", err)
	}

	return SuccessResponse(c, http.StatusOK, "Messages found successfully", messages)
}

// userInChat checks whether the user is in the chat.
func userInChat(user *models.User, chat *models.Chat) bool {
	for _, u := range chat.Users {
		if u.ID == user.ID {
			return true
		}
	}
	return false
}

// getLimit retrieves and validates the 'limit' query parameter.
func getLimit(c echo.Context) (int, error) {
	limit := 10
	limitStr := c.QueryParam("limit")
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 1000 {
			return 0, err
		}
		return limit, nil
	}
	return limit, nil
}

// getFrom retrieves the 'from' query parameter.
func getFrom(c echo.Context) (int, error) {
	from := 0
	fromStr := c.QueryParam("from")
	if fromStr != "" {
		from, err := strconv.Atoi(fromStr)
		if err != nil {
			return 0, err
		}
		return from, err
	}
	return from, nil
}
