package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/peaqnetwork/peaq-network-ev-charging-message-format/golang/message"
)

// Connection holds the live communication between peers
// Events are sent to/from peers
type Connection struct {
	Events chan *message.Event
	ctx    context.Context
	ps     *pubsub.PubSub
	topic  *pubsub.Topic
	sub    *pubsub.Subscription
	redis  *redisServer
	me     peer.ID
	input  chan string
}

// Subscribe tries to subscribe to the PubSub topic to initiate a connection,
// Connection is returned on success.
func Subscribe(ctx context.Context, redis *redisServer, ps *pubsub.PubSub, meID peer.ID, topicName string) (*Connection, error) {
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
		ctx:    ctx,
		ps:     ps,
		redis:  redis,
		topic:  topic,
		sub:    sub,
		me:     meID,
		Events: make(chan *message.Event),
		input:  make(chan string),
	}

	// start reading events from the subscription in a loop
	go conn.fetchEvents()
	return conn, nil
}

func (conn *Connection) ListPeers() []peer.ID {
	return conn.ps.ListPeers(conn.topic.String())
}

// Publish sends a event to peer.
func (conn *Connection) Publish(event string) error {

	msgBytes, err := json.Marshal([]byte(event))
	if err != nil {
		return err
	}
	return conn.topic.Publish(conn.ctx, msgBytes)
}

// fetchEvents pulls events from the subscription connection
// and pushes them onto the Events channel.
func (conn *Connection) fetchEvents() {
	for {
		msg, err := conn.sub.Next(conn.ctx)
		if err != nil {
			close(conn.Events)
			return
		}
		// prevent local peer from sending events itself
		if msg.ReceivedFrom.String() == conn.me.String() {
			continue
		}

		m := new(message.Event)
		err = json.Unmarshal(msg.Data, m)
		if err != nil {
			continue
		}
		// send valid events onto the connection Events channel
		conn.Events <- m
	}
}

func (conn *Connection) ListenEvents() error {
	peerRefreshTicker := time.NewTicker(time.Second)
	defer peerRefreshTicker.Stop()
	peerList := map[string]string{}

ev:
	for {
		select {
		case <-peerRefreshTicker.C:
			peers := conn.ListPeers()
			for _, peer := range peers {
				p := peer.String()
				if _, found := peerList[p]; !found {
					peerList[p] = p
					fmt.Println("New peer connected: ", p)
				}
			}

		case e := <-conn.Events:
			// Display the event received on terminal
			fmt.Println("Event:: ", e)

			bev, err := encode(e)
			if err != nil {
				fmt.Println("Error occurred while publishing to redis:: ", err.Error())
				continue
			}

			// publish to redis channel
			conn.redis.Publish(bev)

		case input := <-conn.input:
			// Publish local peer event
			// Display the event on terminal
			err := conn.Publish(input)
			if err != nil {
				fmt.Printf("publish error: %s", err)
			}
		case <-conn.redis.ctx.Done():
			break ev
		case err := <-conn.redis.done:
			// Return error from the receive goroutine.
			return err
		}
	}

	return nil
}

func encode(ev *message.Event) ([]byte, error) {

	encoded, err := proto.Marshal(ev)

	if err != nil {
		return nil, err
	}

	return encoded, nil
}

func decode(bt []byte) (*message.Event, error) {

	var ev message.Event

	err := proto.Unmarshal(bt, &ev)

	if err != nil {
		return nil, err
	}

	return &ev, nil
}
