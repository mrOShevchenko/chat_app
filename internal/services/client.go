package services

import (
	"chat_app/internal/models"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

const (
	clientsKey  = "clients"
	channelsKey = "channels"
)

// Client represents a chat client
type Client struct {
	User             *models.User
	channelsHandler  *redis.PubSub
	stopListenerChan chan struct{}
	listening        bool
	MessageChan      chan redis.Message
}

// Connect initializes a new chat client and connects it to the client channels on Redis.
func Connect(ctx echo.Context, rdb *redis.Client, user *models.User) (*Client, error) {
	if _, err := rdb.SAdd(ctx.Request().Context(), clientsKey, user.ID).Result(); err != nil {
		return nil, err
	}

	client := &Client{
		User:             user,
		stopListenerChan: make(chan struct{}),
		MessageChan:      make(chan redis.Message),
	}

	if err := client.initializeConnection(ctx, rdb); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) initializeConnection(ctx echo.Context, rdb *redis.Client) error {
	channelsToConnect, err := c.buildChannelsList(ctx, rdb)
	if err != nil {
		return err
	}

	if err := c.disconnectExistingChannels(ctx); err != nil {
		return err
	}

	return c.doConnect(ctx, rdb, channelsToConnect...)
}

func (c *Client) buildChannelsList(ctx echo.Context, rdb *redis.Client) ([]string, error) {
	channelsToConnect, err := rdb.SMembers(ctx.Request().Context(), channelsKey).Result()
	if err != nil {
		return nil, err
	}

	for _, chat := range c.User.Chats {
		channelsToConnect = append(channelsToConnect, fmt.Sprintf("%d", chat.ID))
	}

	return channelsToConnect, nil
}

func (c *Client) disconnectExistingChannels(ctx echo.Context) error {
	if c.channelsHandler != nil {
		if err := c.channelsHandler.Unsubscribe(ctx.Request().Context()); err != nil {
			return err
		}
		if err := c.channelsHandler.Close(); err != nil {
			return err
		}
	}

	if c.listening {
		c.stopListenerChan <- struct{}{}
	}

	return nil
}

func (c *Client) doConnect(ctx echo.Context, rdb *redis.Client, channels ...string) error {
	pubSub := rdb.Subscribe(ctx.Request().Context(), channels...)
	c.channelsHandler = pubSub
	go c.startListener(channels)

	return nil
}

func (c *Client) startListener(channels []string) {
	c.listening = true
	fmt.Printf("Starting the listener for client %d on channels: %s\n", c.User.ID, channels)
	for {
		select {
		case msg, ok := <-c.channelsHandler.Channel():
			if !ok {
				return
			}
			c.MessageChan <- *msg
		case <-c.stopListenerChan:
			fmt.Printf("Stopping the listener for client: %d\n", c.User.ID)
			return
		}
	}
}

// Disconnect closes the client's channels and disconnects the client.
func (c *Client) Disconnect(ctx echo.Context) error {
	if err := c.disconnectExistingChannels(ctx); err != nil {
		return err
	}
	close(c.MessageChan)

	return nil
}

// Chat sends a message to a specific channel.
func Chat(ctx echo.Context, rdb *redis.Client, channel string, content string) error {
	return rdb.Publish(ctx.Request().Context(), channel, content).Err()
}
