// Copyright (C) 2025, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package mdns_test

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/luxfi/mdns"
)

func Example() {
	// Create a discovery instance for your service
	disc := mdns.New("_myservice._tcp", "node-1", 8080,
		mdns.WithLogger(slog.New(slog.NewTextHandler(os.Stdout, nil))),
		mdns.WithMetadata(map[string]string{
			"version": "1.0.0",
			"role":    "worker",
		}),
	)

	// Handle peer events
	disc.OnPeer(func(peer *mdns.Peer, joined bool) {
		if joined {
			fmt.Printf("Peer joined: %s at %s (version: %s)\n",
				peer.NodeID, peer.Address(), peer.Get("version"))
		} else {
			fmt.Printf("Peer left: %s\n", peer.NodeID)
		}
	})

	// Start discovery
	if err := disc.Start(); err != nil {
		fmt.Printf("Failed to start: %v\n", err)
		return
	}
	defer disc.Stop()

	// List all discovered peers
	time.Sleep(100 * time.Millisecond) // Allow time for discovery
	fmt.Printf("Found %d peers\n", disc.PeerCount())
}

func ExampleDiscovery_Peers() {
	disc := mdns.New("_fhed._tcp", "node-1", 8448)

	if err := disc.Start(); err != nil {
		return
	}
	defer disc.Stop()

	// List all peers
	for _, peer := range disc.Peers() {
		fmt.Printf("Node: %s, Address: %s\n", peer.NodeID, peer.Address())
	}
}
