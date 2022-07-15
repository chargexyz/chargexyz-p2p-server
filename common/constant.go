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

	DefaultPort = "10333"

	// ENV VARs
	EnvPort              = "P2P_PORT"
	EnvProviderSecretKey = "PROVIDER_SECRET_KEY"
	EnvRedisHost         = "REDIS_HOST"
	EnvRedisPort         = "REDIS_PORT"
	EnvRedisPubChannel   = "REDIS_PUB_CHANNEL"
	EnvRedisSubChannel   = "REDIS_SUB_CHANNEL"

	// REDIS CONFIGS
	DefaultRedisHost                      = "127.0.0.1"
	DefualtRedisPort                      = "6379"
	DefualtRedisPubChannel                = "IN"
	DefualtRedisSubChannel                = "OUT"
	DefualtRedisChargePointStateKeyPrefix = "CHARGE_POINT_STATE_"

	// A ping is set to the server with this period to test for the health of
	// the connection and server.
	HealthCheckPeriod = time.Minute
)

var (
	EventsPeerCanSend = map[events.EventType]struct{}{
		events.EventType_IDENTITY_CHALLENGE:         {},
		events.EventType_SERVICE_REQUESTED:          {},
		events.EventType_STOP_CHARGE_REQUEST:        {},
		events.EventType_SERVICE_DELIVERY_ACK:       {},
		events.EventType_CHECK_AVAILABILITY_REQUEST: {},
	}

	EventsPeerCanReceive = map[events.EventType]struct{}{
		events.EventType_CHARGING_STATUS:             {},
		events.EventType_IDENTITY_RESPONSE:           {},
		events.EventType_SERVICE_DELIVERED:           {},
		events.EventType_SERVICE_REQUEST_ACK:         {},
		events.EventType_STOP_CHARGE_RESPONSE:        {},
		events.EventType_CHECK_AVAILABILITY_RESPONSE: {},
	}
)
