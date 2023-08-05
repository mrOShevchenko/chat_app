package services

import (
	"chat_app/internal/models"
	"context"
	"encoding/json"
	firebase "firebase.google.com/go"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"strconv"
	"time"
)

// ChatService encapsulates the chat related logic.
type ChatService struct {
	firebaseApp      *firebase.App
	rdb              *redis.Client
	userRepo         models.UserRepository
	messageRepo      models.MessageRepository
	chatRepo         models.ChatRepository
	connectedClients map[int]*Client
}

// msg is the internal message structure used for communication.
type msg struct {
	Content string `json:"content,omitempty"`
	ChatID  int    `json:"chatId,omitempty"`
	Type    string `json:"type,omitempty"` // error, system
	Err     string `json:"error,omitempty"`
}

// NewChatService returns a new instance of ChatService.
func NewChatService(
	firebaseApp *firebase.App,
	rdb *redis.Client,
	userRepo models.UserRepository,
	messageRepo models.MessageRepository,
	chatRepo models.ChatRepository,
) *ChatService {
	return &ChatService{
		firebaseApp:      firebaseApp,
		rdb:              rdb,
		userRepo:         userRepo,
		messageRepo:      messageRepo,
		chatRepo:         chatRepo,
		connectedClients: make(map[int]*Client),
	}
}

// OnConnect manages the actions required when a user connects.
func (cs *ChatService) OnConnect(ctx context.Context, conn *websocket.Conn, user *models.User) error {
	log.Printf("user %s id: %d, connected from: %s \n", user.Username, user.ID, conn.RemoteAddr())

	u, err := Connect(ctx, cs.rdb, user)
	if err != nil {
		return err
	}
	cs.connectedClients[user.ID] = u
	return nil
}

// OnDisconnect manages the actions required when a user disconnects.
func (cs *ChatService) OnDisconnect(ctx context.Context, conn *websocket.Conn, user *models.User) chan struct{} {
	closeCh := make(chan struct{})

	conn.SetCloseHandler(func(code int, text string) error {
		log.Printf("connection closed for client %s id: %d", user.Username, user.ID)

		u := cs.connectedClients[user.ID]
		if err := u.Disconnect(ctx); err != nil {
			return err
		}
		delete(cs.connectedClients, user.ID)
		close(closeCh)
		return nil
	})

	return closeCh
}

// OnClientMessage manages the actions required when a client message is received.
func (cs *ChatService) OnClientMessage(ctx context.Context, conn *websocket.Conn, user *models.User) {
	msg, err := cs.readMessage(conn)
	if err != nil {
		cs.HandleWSError(err, "error reading message", conn)
		return
	}
	log.Printf("message: %s, client: %s id:%d ", msg.Content, user.Username, user.ID)

	newMessage, err := cs.createNewMessage(msg, user)
	if err != nil {
		cs.HandleWSError(err, "error reading message", conn)
		return
	}

	chat, err := cs.getChat(msg)
	if err != nil {
		cs.HandleWSError(err, "error marshaling new message", conn)
		return
	}

	if cs.isUserBlockedInChat(chat, user) {
		return
	}

	err = cs.sendMessage(ctx, msg, newMessage)
	if err != nil {
		cs.HandleWSError(err, "error sending message to channel", conn)
		return
	}

	err = cs.saveMessage(newMessage)
	if err != nil {
		log.Println(errors.Wrap(err, "can't create new message in repository"))
	}
}

func (cs *ChatService) readMessage(conn *websocket.Conn) (msg, error) {
	var msg msg
	err := conn.ReadJSON(&msg)
	if err != nil {
		return msg, err
	}
	return msg, nil
}

func (cs *ChatService) createNewMessage(msg msg, user *models.User) (models.Message, error) {
	newMessage := models.Message{
		Type:      "text",
		ChatID:    msg.ChatID,
		Content:   msg.Content,
		SenderID:  user.ID,
		Sender:    *user,
		Status:    "sent",
		CreatedAt: time.Now().Unix(),
	}

	_, err := json.Marshal(newMessage)
	if err != nil {
		return newMessage, err
	}

	return newMessage, nil
}

func (cs *ChatService) getChat(msg msg) (*models.Chat, error) {
	chat, err := cs.chatRepo.FindByID(msg.ChatID)
	if err != nil {
		return chat, err
	}

	return chat, nil
}

func (cs *ChatService) isUserBlockedInChat(chat *models.Chat, user *models.User) bool {
	if chat.Type == "private" {
		for _, u := range chat.Users {
			if u.ID != user.ID {
				for _, blocked := range u.BlockedUsers {
					if user.ID == blocked.ID {
						log.Printf("user %d blocks %d", u.ID, user.ID)
						return true
					}
				}
			}
		}
	}

	return false
}

func (cs *ChatService) sendMessage(ctx context.Context, msg msg, newMessage models.Message) error {
	channelID := strconv.Itoa(msg.ChatID)

	err := Chat(ctx, cs.rdb, channelID, newMessage.Content)
	if err != nil {
		return err
	}

	log.Printf("message: %s,- sent to channel: %s", newMessage.Content, channelID)
	return nil
}

func (cs *ChatService) saveMessage(newMessage models.Message) error {
	return cs.messageRepo.Create(&newMessage)
}

// OnChannelMessage manages the actions required when a channel message is received.
func (cs *ChatService) OnChannelMessage(ctx context.Context, conn *websocket.Conn, user *models.User) {
	c := cs.connectedClients[user.ID]

	go func() {
		for m := range c.MessageChan {
			var message models.Message
			err := json.Unmarshal([]byte(m.Payload), &message)
			if err != nil {
				cs.HandleWSError(err, "error unmarshaling channel message", conn)
				continue
			}

			if message.SenderID != user.ID {
				err = conn.WriteJSON(message)
				if err != nil {
					log.Println(errors.Wrap(err, "error writing message to connection"))
					continue
				}
				cs.SendNotification(ctx, &message)
			}
		}
	}()
}

// HandleWSError handles websocket errors.
func (cs *ChatService) HandleWSError(err error, message string, conn *websocket.Conn) {
	payload := msg{
		Type:    "error",
		Content: message,
	}

	if os.Getenv("APP_ENV") == "dev" {
		payload.Err = err.Error()
	}

	_ = conn.WriteJSON(payload)
}
