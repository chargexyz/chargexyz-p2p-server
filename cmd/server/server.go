package server

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"flag"
	"fmt"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/peaqnetwork/peaq-network-ev-charging-sim-be-p2p/common"
	"github.com/peaqnetwork/peaq-network-ev-charging-sim-be-p2p/services"
)

// Used for checks
var localPeerID peer.ID

func Run() error {

	// parse required flags
	portFlag := flag.String("p", "0", "listening port")
	skFlag := flag.String("sk", "", "Hex string of the provider secret key/seed")
	flag.Parse()

	port := *portFlag
	sk := *skFlag

	prvKey, err := generatePrivateKey(sk)
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

	// subscribe to the topic
	conn, err := services.Subscribe(ctx, ps, localPeerID, common.TOPIC)
	if err != nil {
		return err
	}
	fmt.Println("Local Peer ID", localPeerID)
	fmt.Println("Listening on...", h.Addrs())

	//standard input for user message entry in terminal
	go conn.WriteMessage()
	conn.ListenEvents()

	return nil
}

// Generates ED25519 private key
func generatePrivateKey(sk string) (crypto.PrivKey, error) {
	seed, err := hex.DecodeString(sk)
	if err != nil {
		return nil, fmt.Errorf("unable to parse provided secret key: %s", err.Error())
	}

	key := ed25519.NewKeyFromSeed(seed)
	prvKey, err := crypto.UnmarshalEd25519PrivateKey(key)
	if err != nil {
		return prvKey, fmt.Errorf("unable to parse provided secret key: %s", err.Error())
	}

	return prvKey, nil
}
