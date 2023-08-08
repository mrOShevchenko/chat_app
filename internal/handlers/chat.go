package handlers

import (
	"chat_app/internal/models"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
	"reflect"
	"strconv"
)

type MessageResponseBody struct {
	Content     string `json:"content" example:"twit-twit"`
	RecipientID int    `json:"recipientId" example:"1"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     CheckOrigin,
}

// Chat godoc
// @Summary Chat [WebSocket]
// @Description Chat with users based on WebSocket
// @Tags chat
// @Param token query string true "Access JWT Token"
// @Param chat body MessageResponseBody true "body should contain content and recipient_id for sending message"
// @Failure 401 {object} Response
// @Router /ws/chat [get]
func (h *BaseHandler) Chat(c echo.Context) error {
	token := c.QueryParam("token")
	accessTokenClaims, err := h.tokenService.DecodeAccessToken(token)
	if err != nil {
		return ErrorResponse(c, http.StatusUnauthorized, "invalid token", err)
	}

	ctx := c.Request().Context()
	_, err = h.tokenService.GetCacheValue(ctx, accessTokenClaims.AccessUUID)
	if err != nil {
		return ErrorResponse(c, http.StatusUnauthorized, "invalid token", err)
	}

	user, err := h.userRepo.FindByID(accessTokenClaims.UserID)
	if err != nil {
		return ErrorResponse(c, http.StatusUnauthorized, "user was not found", err)
	}

	c.Logger().Info(fmt.Sprintf("client %d connected", user.ID))
	user.IsOnline = true
	if err = h.userRepo.Update(user); err != nil {
		c.Logger().Error(err.Error())
		return ErrorResponse(c, http.StatusInternalServerError, "failed to update user status", err)
	}

	// Setting up WebSocket
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "failed to upgrade to websocket", err)
	}
	defer func(ws *websocket.Conn) {
		err := ws.Close()
		if err != nil {
			c.Logger().Errorf("error closing websocket: ", err)
		}
	}(ws)

	if err = h.chatService.OnConnect(ctx, ws, user); err != nil {
		h.chatService.HandleWSError(err, "error sending message", ws)
		return err
	}

	closeCh := h.chatService.OnDisconnect(ctx, ws, user)

	h.chatService.OnChannelMessage(ctx, ws, user)

loop:
	for {
		select {
		case <-closeCh:
			user.IsOnline = false
			if err := h.userRepo.Update(user); err != nil {
				c.Logger().Error(err.Error())
				return ErrorResponse(c, http.StatusInternalServerError, "failed to update user status", err)
			}
			break loop
		default:
			h.chatService.OnClientMessage(ctx, ws, user)
		}
	}
	return nil
}

// CreateChat godoc
// @Summary Create Chat by User ID
// @Tags chat
// @Accept  json
// @Produce application/json
// @Param	user_id	path	int	true	"User ID"
// @Security BearerAuth
// @Success 200 {object} Response{data=models.Chat}
// @Failure 400 {object} Response
// @Failure 409 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/{user_id}/chat [post]
func (h *BaseHandler) CreateChat(c echo.Context) error {
	user := c.Get("user").(*models.User)
	chatUserID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid user id", err)
	}

	if user.ID == chatUserID {
		return ErrorResponse(c, http.StatusBadRequest, "can't create chat with himself", err)
	}

	chatUser, err := h.userRepo.FindByID(chatUserID)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "user was not found", err)
	}
	if chatUser == nil || reflect.DeepEqual(*chatUser, models.User{}) {
		return ErrorResponse(c, http.StatusBadRequest, "user not found", err)
	}

	chat, err := h.chatRepo.FindPrivateChatByUsersArray([]*models.User{user, chatUser})
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrorResponse(c, http.StatusInternalServerError, "error while creating chat", err)
	}

	if chat != nil {
		return ErrorResponse(c, http.StatusConflict, "chat already exists", err)
	}

	var newChat = models.Chat{
		Type:  "private",
		Users: []*models.User{user, chatUser},
	}
	if err = h.chatRepo.Create(&newChat); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "error while creating chat", err)
	}
	return SuccessResponse(c, http.StatusCreated, "chat created successfully", newChat)
}

// GetChat godoc
// @Summary Get Chat by User ID
// @Tags chat
// @Accept  json
// @Produce application/json
// @Param	user_id	path	int	true	"User ID"
// @Security BearerAuth
// @Success 200 {object} Response{data=models.Chat}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /v1/user/{user_id}/chat [get]
func (h *BaseHandler) GetChat(c echo.Context) error {
	user := c.Get("user").(*models.User)
	chatUserID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid user id", err)
	}
	if user.ID == chatUserID {
		err := fmt.Errorf("user can't create chat with himself")
		return ErrorResponse(c, http.StatusBadRequest, "user can't create chat with himself", err)
	}

	chatUser, err := h.userRepo.FindByID(chatUserID)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid user", err)
	}
	if chatUser == nil || reflect.DeepEqual(*chatUser, models.User{}) {
		err := fmt.Errorf("user with id %d not found", chatUserID)
		return ErrorResponse(c, http.StatusBadRequest, "user not found", err)
	}

	chat, err := h.chatRepo.FindPrivateChatByUsersArray([]*models.User{user, chatUser})
	fmt.Println(chat)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrorResponse(c, http.StatusInternalServerError, "error while searching chat", err)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrorResponse(c, http.StatusNotFound, "chat was not found", err)
	}
	if reflect.DeepEqual(*chat, models.Chat{}) {
		err := fmt.Errorf("private chat between %d & %d Users not found", user.ID, chatUser.ID)
		return ErrorResponse(c, http.StatusNotFound, "chat was not found", err)
	}
	return SuccessResponse(c, http.StatusOK, "chat successfully found", chat)
}

// DeleteChat godoc
// @Summary Delete Chat by ChatID
// @Tags chat
// @Accept  json
// @Produce application/json
// @Param	chat_id	path	int	true	"Chat ID"
// @Security BearerAuth
// @Success 200 {object} Response{data=models.User}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 403 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /v1/chat/{chat_id} [delete]
func (h *BaseHandler) DeleteChat(c echo.Context) error {
	user := c.Get("user").(*models.User)
	rawData := c.Param("chat_id")
	chatID, err := strconv.Atoi(rawData)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "invalid chat's userId", err)
	}

	chat, err := h.chatRepo.FindByID(chatID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrorResponse(c, http.StatusInternalServerError, "error while deleting chat", err)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrorResponse(c, http.StatusNotFound, "chat not found", err)
	}

	if chat == nil || reflect.DeepEqual(*chat, models.Chat{}) {
		err := fmt.Errorf("private chat between %d not found", chatID)
		return ErrorResponse(c, http.StatusNotFound, "chat not found", err)
	}

	if chat.Type != "private" {
		return ErrorResponse(c, http.StatusForbidden, "can't delete this chat: have not access", nil)
	}

	for _, item := range chat.Users {
		if item.ID == user.ID {
			err := h.chatRepo.Delete(chat)
			if err != nil {
				return ErrorResponse(c, http.StatusInternalServerError, "failed to delete chat", err)
			}

			return SuccessResponse(c, http.StatusOK, "successfully deleted chat", nil)
		}
	}

	err = fmt.Errorf("have not access to delete this chat")
	return ErrorResponse(c, http.StatusForbidden, "can't delete this chat: have not access", err)
}
