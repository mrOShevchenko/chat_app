package services

import (
	"chat_app/internal/models"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

const (
	clientsKey  = "clients"
	channelsKey = "channels"
)

type Client struct {
	User            *models.User
	channelsHandler *redis.PubSub

	stopListenerChan chan struct{}
	listening        bool

	MessageChan chan redis.Message
}

// Connect client to client channels on redis
func Connect(rdb *redis.Client, user *models.User) (*Client, error) {
	if _, err := rdb.SAdd(context.Background(), clientsKey, user.ID).Result(); err != nil {
		return nil, err
	}

	c := &Client{
		User:             user,
		stopListenerChan: make(chan struct{}),
		MessageChan:      make(chan redis.Message),
	}

	if err := c.connect(rdb); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) connect(rdb *redis.Client) error {

	var c0 []string

	c1, err := rdb.SMembers(context.Background(), channelsKey).Result()
	if err != nil {
		return err
	}

	c0 = append(c0, c1...)

	for _, chat := range c.User.Chats {
		c0 = append(c0, fmt.Sprintf("%d", chat.ID))
	}

	if len(c0) == 0 {
		fmt.Println("no channels to connect to client: ", c.User.ID)
		return nil
	}

	if c.channelsHandler != nil {
		if err = c.channelsHandler.Unsubscribe(context.Background()); err != nil {
			return err
		}
		if err = c.channelsHandler.Close(); err != nil {
			return err
		}
	}
	if c.listening {
		c.stopListenerChan <- struct{}{}
	}

	return c.doConnect(rdb, c0...)
}

func (c *Client) doConnect(rdb *redis.Client, channels ...string) error {
	//subcsribe all channels in one request
	pubSub := rdb.Subscribe(context.Background(), channels...)
	// keep channel handler to be used in unsubscribe
	c.channelsHandler = pubSub

	// the Listener
	go func() {
		c.listening = true
		fmt.Printf("starting the listener for client %d on channels: %s", c.User.ID, channels)
		for {
			select {
			case msg, ok := <-pubSub.Channel():
				if !ok {
					return
				}
				c.MessageChan <- *msg
			case <-c.stopListenerChan:
				fmt.Printf("stopping the listener for client: %d", c.User.ID)
				return
			}
		}

	}()
	return nil
}

func (c *Client) Disconnect() error {
	if c.channelsHandler != nil {
		if err := c.channelsHandler.Unsubscribe(context.Background()); err != nil {
			return err
		}
		if err := c.channelsHandler.Close(); err != nil {
			return err
		}
	}
	if c.listening {
		c.stopListenerChan <- struct{}{}
	}

	close(c.MessageChan)

	return nil
}

func Chat(rdb *redis.Client, channel string, content string) error {
	return rdb.Publish(context.Background(), channel, content).Err()
}
