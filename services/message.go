package services

import (
	"context"
	"encoding/json"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/peaqnetwork/peaq-network-ev-charging-sim-be-p2p/common"
)

// Connection holds the live communication between peers
// Messages are sent to/from peers
type Connection struct {
	// Messages is a channel of messages received from peer
	Messages chan *Message
	ctx      context.Context
	ps       *pubsub.PubSub
	topic    *pubsub.Topic
	sub      *pubsub.Subscription
	me       peer.ID
}

// Message gets converted to/from JSON and sent in the body of pubsub messages.
type Message struct {
	Message string
	Sender  string
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

	cr := &Connection{
		ctx:      ctx,
		ps:       ps,
		topic:    topic,
		sub:      sub,
		me:       meID,
		Messages: make(chan *Message, common.ConnectionBufSize),
	}

	// start reading messages from the subscription in a loop
	go cr.fetchMessages()
	return cr, nil
}

// fetchMessages pulls messages from the pubsub topic and pushes them onto the Messages channel.
func (conn *Connection) fetchMessages() {
	for {
		msg, err := conn.sub.Next(conn.ctx)
		if err != nil {
			close(conn.Messages)
			return
		}
		// prevent me from sending message to self
		if msg.ReceivedFrom.String() == conn.me.String() {
			continue
		}

		cm := new(Message)
		err = json.Unmarshal(msg.Data, cm)
		if err != nil {
			continue
		}
		// send valid messages onto the Messages channel
		conn.Messages <- cm
	}
}
