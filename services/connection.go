package services

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
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
	sk     ed25519.PrivateKey
}

// Subscribe tries to subscribe to the PubSub topic to initiate a connection,
// Connection is returned on success.
func Subscribe(ctx context.Context, redis *redisServer, ps *pubsub.PubSub, meID peer.ID, topicName string, signKey ed25519.PrivateKey) (*Connection, error) {
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
		sk:     signKey,
	}

	// start reading events from the subscription in a loop
	go conn.fetchEvents()
	return conn, nil
}

func (conn *Connection) ListPeers() []peer.ID {
	return conn.ps.ListPeers(conn.topic.String())
}

// Publish sends a event to peer.
func (conn *Connection) Publish(ev *message.Event) error {

	fmt.Println("Publish Event ", ev)

	msgBytes, err := encode(ev)
	if err != nil {
		return err
	}
	fmt.Println("Publish Event bytes ", msgBytes)
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

		ev, err := decode(msg.Data)
		if err != nil {
			fmt.Println("Error occurred while parsing protobuf:: ", err)
			continue
		}
		// send valid events onto the connection Events channel
		conn.Events <- ev
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
			fmt.Println("Event ID:: ", e.EventId)
			fmt.Println("Event Data:: ", e.Data)

			if e.EventId == message.EventType_IDENTITY_CHALLENGE {
				conn.parseIdentityChallenge(e)
				continue
			}

			evHexString, err := encodeToHex(e)
			if err != nil {
				fmt.Println("Error occurred while publishing to redis:: ", err.Error())
				continue
			}

			// publish to redis channel
			conn.redis.Publish(evHexString)

		case input := <-conn.redis.input:
			// Publish redis event to peer
			ev, err := decodeFromHex(input)
			if err != nil {
				fmt.Printf("error decode redis hex string. error: %s", err)
			}

			err = conn.Publish(ev)
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

// parseIdentityChallenge - takes the plain data sent by consumer and sign it with private key
func (conn *Connection) parseIdentityChallenge(ev *message.Event) {
	plainData := ev.GetIdentityChallengeData().PlainData
	hsh := ed25519.Sign(conn.sk, []byte(plainData))
	encodedString := hex.EncodeToString(hsh)

	fmt.Println("Encoded hsh:: ", encodedString)
	response := message.Event{
		EventId: message.EventType_IDENTITY_RESPONSE,
		Data: &message.Event_IdentityResponseData{
			IdentityResponseData: &message.IdentityResponseData{
				Resp: &message.Response{
					Error:   false,
					Message: "success",
				},
				Signature: encodedString,
			},
		},
	}

	conn.Publish(&response)

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

// encodeToHex is used when sending event to redis
func encodeToHex(ev *message.Event) (string, error) {

	encoded, err := encode(ev)
	if err != nil {
		return "", err
	}
	// Encode the byte to hex string
	encodedString := hex.EncodeToString(encoded)

	return encodedString, nil
}

// decodeFromHex is used when reading redis event
func decodeFromHex(bt []byte) (*message.Event, error) {

	// convert the byte to hexstring
	hexString := string(bt)

	// decode the hex string to byte slice
	bytt, err := hex.DecodeString(hexString)
	if err != nil {
		return nil, err
	}

	// decode to Event
	ev, err := decode(bytt)
	if err != nil {
		return nil, err
	}

	return ev, nil
}
