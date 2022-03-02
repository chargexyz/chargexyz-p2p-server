package common

import "time"

// ConnectionBufSize is the number of incoming messages to buffer.
const ConnectionBufSize = 128

// DiscoveryInterval is how often we re-publish our mDNS records.
const DiscoveryInterval = time.Hour

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DiscoveryServiceTag = "charmev"

// Subscription topic
const TOPIC = "charmev"
