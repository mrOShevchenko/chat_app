package main

import (
	"chat_app/configs"
	"chat_app/internal"
	"chat_app/internal/handlers"
	"chat_app/internal/models"
	"chat_app/internal/services"
	"chat_app/internal/services/tokenService"
	"chat_app/internal/storage"
	firebaseClient "chat_app/pkg/firebase_client"
	gormClient "chat_app/pkg/gorm_client"
	redisClient "chat_app/pkg/redis_client"
	"fmt"
	"github.com/labstack/gommon/log"

)

const webPort = 80

func main() {
	clientGORM, err := gormClient.NewClient()
	if err != nil {
		log.Fatalf("error with creating New Gorm Client: %s", err)
	}
	err = clientGORM.AutoMigrate(models.User{}, models.Chat{}, models.Device{})
	if err != nil {
		log.Fatalf("error Automigrate: %s", err)
	}

	ctx, firebaseApp, _, err := firebaseClient.SetupFirebase()
	if err != nil {
		log.Printf("error setup firebase: %s", err)
	}

	clientREDIS, err := redisClient.NewClient(ctx)
	if err != nil {
		log.Fatalf("error creating redis client: %s", err)
	}

	userRepo := storage.NewUserRepo(clientGORM, clientREDIS)
	messageRepo := storage.NewMessageRepo(clientGORM)
	chatRepo := storage.NewChatRepo(clientGORM)

	tokenServ := tokenService.NewService(clientREDIS)
	chatService := services.NewChatService(firebaseApp, clientREDIS, userRepo, messageRepo, chatRepo)

	baseHandler := handlers.NewBaseHandler(userRepo, messageRepo, chatRepo, tokenServ, chatService)

	config := &configs.Config{
		BaseHandler:  baseHandler,
		UserRepo:     userRepo,
		TokenService: tokenServ,
	}

	appConfig := internal.AppConfig{
		config,
	}

	e := appConfig.NewRouter()
	appConfig.AddMiddleware(e)
	e.Logger.SetLevel(log.INFO)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", webPort)))
}
