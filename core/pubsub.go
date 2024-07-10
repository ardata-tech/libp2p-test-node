package core

import (
	"context"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"log"
)

type PubSub struct {
	PS    *pubsub.PubSub
	Topic *pubsub.Topic
	Sub   *pubsub.Subscription
}

func NewPubSub(h host.Host, topicName string) *PubSub {
	ps, err := pubsub.NewGossipSub(context.Background(), h)
	if err != nil {
		log.Fatal(err)
	}

	topic, err := ps.Join(topicName)
	if err != nil {
		log.Fatal(err)
	}

	sub, err := topic.Subscribe()
	if err != nil {
		log.Fatal(err)
	}
	return &PubSub{
		PS:    ps,
		Topic: topic,
		Sub:   sub,
	}
}
