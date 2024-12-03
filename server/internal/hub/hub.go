package hub

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Hub struct {
	Client      *redis.Client
	ChannelName string
}

func (h *Hub) Publish(ctx context.Context, mesg string) error {
	return h.Client.Publish(ctx, h.ChannelName, mesg).Err()
}

func (h *Hub) Subscribe(ctx context.Context) (*Subscriber, error) {
	pubsub := h.Client.Subscribe(ctx, h.ChannelName)
	return &Subscriber{pubsub}, nil
}

type Subscriber struct {
	pubSub *redis.PubSub
}

func (s *Subscriber) Channel() <-chan string {
	ch := make(chan string)
	go func() {
		for mesg := range s.pubSub.Channel() {
			ch <- string(mesg.Payload)
		}
	}()
	return ch
}

func (s *Subscriber) Close() error {
	return s.pubSub.Close()
}
