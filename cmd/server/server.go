package server

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/sirupsen/logrus"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/peaqnetwork/peaq-network-ev-charging-sim-be-p2p/common"
	"github.com/peaqnetwork/peaq-network-ev-charging-sim-be-p2p/services"
)

// Used for checks
var localPeerID peer.ID

func Run() error {

	sk := os.Getenv(common.EnvProviderSecretKey)
	if len(sk) < 1 {
		logrus.Panicln(common.EnvProviderSecretKey, " env var not set")
	}
	if len(sk) < 64 {
		logrus.Panicln("invalid env var value: ", common.EnvProviderSecretKey)
	}

	port := os.Getenv(common.EnvPort)
	if len(port) < 1 {
		logrus.Infof("%s env var not set. using default port: %v", common.EnvPort, common.DefaultPort)
		port = common.DefaultPort
	}

	redisHost := os.Getenv(common.EnvRedisHost)
	if len(redisHost) < 1 {
		logrus.Infof("%s env var not set. using default redis host: %v", common.EnvRedisHost, common.DefaultRedisHost)
		redisHost = common.DefaultRedisHost
	}

	redisPort := os.Getenv(common.EnvRedisPort)
	if len(redisPort) < 1 {
		logrus.Infof("%s env var not set. using default redis port: %v", common.EnvRedisPort, common.DefualtRedisPort)
		redisPort = common.DefualtRedisPort
	}

	redisPubChannel := os.Getenv(common.EnvRedisPubChannel)
	if len(redisPubChannel) < 1 {
		logrus.Infof("%s env var not set. using default redis pub channel: %v", common.EnvRedisPubChannel, common.DefualtRedisPubChannel)
		redisPubChannel = common.DefualtRedisPubChannel
	}

	redisSubChannel := os.Getenv(common.EnvRedisSubChannel)
	if len(redisSubChannel) < 1 {
		logrus.Infof("%s env var not set. using default redis sub channel: %v", common.EnvRedisSubChannel, common.DefualtRedisSubChannel)
		redisSubChannel = common.DefualtRedisSubChannel
	}

	prvKey, signkey, err := generatePrivateKey(sk)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", port)

	// create a new libp2p Host that listens on a random/provided TCP port
	h, err := libp2p.New(libp2p.ListenAddrStrings(addr),
		libp2p.Identity(prvKey))
	if err != nil {
		return err
	}

	localPeerID = h.ID()
	ctx := context.Background()

	// create a new PubSub service using the GossipSub router
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return err
	}

	// setup local mDNS discovery
	if err := setupDiscovery(h); err != nil {
		return err
	}

	// Connect to redis server
	config := services.Config{
		Host:       redisHost,
		Port:       redisPort,
		PubChannel: redisPubChannel,
		SubChannel: redisSubChannel,
	}
	redis := services.NewRedisServer(config)
	err = redis.Run(ctx)
	if err != nil {
		return err
	}

	// subscribe to the topic
	conn, err := services.Subscribe(ctx, redis, ps, localPeerID, common.TOPIC, signkey)
	if err != nil {
		return err
	}

	fmt.Println("Local Peer ID", localPeerID)
	fmt.Println("Listening on...", h.Addrs())

	if err := conn.ListenEvents(); err != nil {
		return err
	}

	return nil
}

// Generates ED25519 private key
func generatePrivateKey(sk string) (crypto.PrivKey, ed25519.PrivateKey, error) {
	seed, err := hex.DecodeString(sk)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse provided secret key: %s", err.Error())
	}

	key := ed25519.NewKeyFromSeed(seed)

	prvKey, err := crypto.UnmarshalEd25519PrivateKey(key)
	if err != nil {
		return prvKey, key, fmt.Errorf("unable to parse provided secret key: %s", err.Error())
	}

	return prvKey, key, nil
}
