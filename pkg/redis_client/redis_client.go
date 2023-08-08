package redisClient

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
	"log"
	"os"
	"time"
)

// Maximum number of connection attempts
const maxAttempts = 10

// NewClient creates a new Redis client.
// It keeps attempting to connect to the database until successful, or until maxAttempts has been reached.
// In case of a failure, it returns an error.
func NewClient(ctx context.Context) (*redis.Client, error) {
	dsn := os.Getenv("REDIS_DSN")
	password := os.Getenv("REDIS_PASSWORD")

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		client := redis.NewClient(&redis.Options{
			Addr:     dsn,
			Password: password,
		})

		_, err := client.Ping(ctx).Result()

		if err != nil {
			log.Printf("Attempt %d: Redis not ready, reconnect after two seconds...", attempt)
			time.Sleep(2 * time.Second)
		} else {
			log.Println("Connected to Redis!")
			return client, nil
		}
	}

	return nil, fmt.Errorf("connection attempts exceeded: could not connect to Redis")
}
