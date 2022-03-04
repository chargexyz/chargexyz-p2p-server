package services

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

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
	redis    *redisServer
	me       peer.ID
	Input    chan string
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
		ctx:      ctx,
		ps:       ps,
		redis:    redis,
		topic:    topic,
		sub:      sub,
		me:       meID,
		Messages: make(chan *Message),
		Input:    make(chan string),
	}

	// start reading messages from the subscription in a loop
	go conn.fetchMessages()
	return conn, nil
}

func (conn *Connection) ListPeers() []peer.ID {
	return conn.ps.ListPeers(conn.topic.String())
}

// Publish sends a message to peer.
func (conn *Connection) Publish(message string) error {
	m := Message{
		Message: message,
		Sender:  conn.me.String(),
	}
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return conn.topic.Publish(conn.ctx, msgBytes)
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

func (conn *Connection) WriteMessage() {
	io := bufio.NewReader(os.Stdin)
	for {
		in, _ := io.ReadString('\n')
		in = strings.TrimSuffix(in, "\n")
		in = strings.Trim(in, " ")
		conn.Input <- in
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

		case m := <-conn.Messages:
			// Display the message received on terminal
			fmt.Println(m.Sender, " > ", m.Message)

		case input := <-conn.Input:
			// Publish local peer message
			// Display the message on terminal
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
