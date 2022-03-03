package services

import (
	"context"
	"encoding/json"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// Connection holds the live communication between peers
// Messages are sent to/from peers
type Connection struct {
	Messages chan *Message
	ctx      context.Context
	ps       *pubsub.PubSub
	topic    *pubsub.Topic
	sub      *pubsub.Subscription
	me       peer.ID
	done     chan struct{}
}

// Subscribe tries to subscribe to the PubSub topic to initiate a connection,
// Connection is returned on success.
func Subscribe(ctx context.Context, ps *pubsub.PubSub, meID peer.ID, topicName string) (*Connection, error) {
	// join the pubsub topic
	topic, err := ps.Join(topicName)
	if err != nil {
		return nil, err
	}

	// and subscribe to it
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		ctx:      ctx,
		ps:       ps,
		topic:    topic,
		sub:      sub,
		me:       meID,
		Messages: make(chan *Message),
	}

	// start reading messages from the subscription in a loop
	go conn.fetchMessages()
	return conn, nil
}

func (conn *Connection) ListPeers() []peer.ID {
	return conn.ps.ListPeers(conn.topic.String())
}

// fetchMessages pulls messages from the subscription connection
// and pushes them onto the Messages channel.
func (conn *Connection) fetchMessages() {
	for {
		msg, err := conn.sub.Next(conn.ctx)
		if err != nil {
			close(conn.Messages)
			return
		}
		// prevent local peer from sending messages itself
		if msg.ReceivedFrom.String() == conn.me.String() {
			continue
		}

		m := new(Message)
		err = json.Unmarshal(msg.Data, m)
		if err != nil {
			continue
		}
		// send valid messages onto the connection Messages channel
		conn.Messages <- m
	}
}
