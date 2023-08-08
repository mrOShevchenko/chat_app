package firebaseClient

import (
	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"log"
	"os"
	"path/filepath"
)

// SetupFirebase initializes Firebase with the given service account key,
// it returns the Firebase app instance, a context, and a messaging client.
func SetupFirebase() (context.Context, *firebase.App, *messaging.Client, error) {
	ctx := context.Background()
	serviceAccountKeyFilePath, err := filepath.Abs(os.Getenv("FIREBASE_KEY_PATH"))
	if err != nil {
		log.Printf("Unable to load serviceAccountKeys.json file: %v", err)
		return ctx, nil, nil, err
	}
	opt := option.WithCredentialsFile(serviceAccountKeyFilePath)

	// Firebase admin SDK initialization
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Printf("Error initializing Firebase app: %v", err)
		return ctx, nil, nil, err
	}

	// Messaging client
	client, err := app.Messaging(ctx)
	if err != nil {
		log.Printf("Error initializing Firebase messaging client: %v", err)
		return ctx, nil, nil, err
	}

	return ctx, app, client, nil
}
