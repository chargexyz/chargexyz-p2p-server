package common

import (
	"time"

	"github.com/peaqnetwork/peaq-network-ev-charging-message-format/golang/events"
)

const (
	// ConnectionBufSize is the number of incoming messages to buffer.
	ConnectionBufSize = 128

	// DiscoveryInterval is how often we re-publish our mDNS records.
	DiscoveryInterval = time.Hour

	// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
	DiscoveryServiceTag = "charmev"

	// Subscription topic
	TOPIC = "charmev"

	// REDIS CONFIGS
	Host       = "127.0.0.1"
	Port       = "6379"
	PubChannel = "IN"
	SubChannel = "OUT"
	// A ping is set to the server with this period to test for the health of
	// the connection and server.
	HealthCheckPeriod = time.Minute
)

var (
	EventsPeerCanSend = map[events.EventType]struct{}{
		events.EventType_IDENTITY_CHALLENGE:   {},
		events.EventType_SERVICE_REQUESTED:    {},
		events.EventType_STOP_CHARGE_REQUEST:  {},
		events.EventType_SERVICE_DELIVERY_ACK: {},
	}

	EventsPeerCanReceive = map[events.EventType]struct{}{
		events.EventType_CHARGING_STATUS:      {},
		events.EventType_IDENTITY_RESPONSE:    {},
		events.EventType_SERVICE_DELIVERED:    {},
		events.EventType_SERVICE_REQUEST_ACK:  {},
		events.EventType_STOP_CHARGE_RESPONSE: {},
	}
)
