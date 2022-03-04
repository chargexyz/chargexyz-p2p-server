package common

import "time"

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
