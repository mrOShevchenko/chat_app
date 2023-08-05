package services

import (
	"chat_app/internal/models"
	"firebase.google.com/go/messaging"
	"fmt"
	"golang.org/x/net/context"
	"log"
)

// Fetches the chat based on message.ChatID
func (cs *ChatService) fetchChatFromMessage(message *models.Message) (*models.Chat, error) {
	chat, err := cs.chatRepo.FindByID(message.ChatID)
	if err != nil {
		log.Printf("error with getting chat: %s", err)
		return nil, err
	}
	return chat, nil
}

// Fetches the user based on user ID
func (cs *ChatService) fetchUserByID(userID int) (*models.User, error) {
	user, err := cs.userRepo.FindByID(userID)
	if err != nil {
		log.Printf("error getting user: %s", err)
		return nil, err
	}
	return user, nil
}

// Sends a message to a user device
func (cs *ChatService) sendMessageToDevice(ctx context.Context, user *models.User, message *models.Message) {
	client, err := cs.firebaseApp.Messaging(ctx)
	if err != nil {
		log.Printf("error getting Messaging client: %v \n", err)
		return
	}

	for _, d := range user.Devices {
		msg := &messaging.Message{
			Notification: &messaging.Notification{
				Title: fmt.Sprintf("Message from %s", message.Sender.Username),
				Body:  message.Content,
			},
			Token: d.Token,
		}

		response, err := client.Send(ctx, msg)
		if err != nil {
			log.Printf("error sending message: %s \n", err)
			continue
		}

		log.Printf("successfully sent message: %s", response)
	}
}

// SendNotification sends notifications to users.
func (cs *ChatService) SendNotification(ctx context.Context, message *models.Message) {
	chat, err := cs.fetchChatFromMessage(message)
	if err != nil {
		return
	}

	for _, u := range chat.Users {
		if u.ID == message.SenderID {
			continue
		}

		user, err := cs.fetchUserByID(u.ID)
		if err != nil {
			return
		}

		// Handle case when user has devices.
		if len(user.Devices) > 0 {
			cs.sendMessageToDevice(ctx, user, message)
		}
	}
}
