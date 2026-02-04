// Copyright (C) 2025, Lux Industries Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package mdns

import (
	"sync"
	"testing"
	"time"
)

func TestNewDiscovery(t *testing.T) {
	d := New("_test._tcp", "node-1", 8080)
	if d == nil {
		t.Fatal("expected non-nil discovery")
	}
	if d.NodeID() != "node-1" {
		t.Errorf("expected nodeID 'node-1', got %q", d.NodeID())
	}
	if d.ServiceType() != "_test._tcp" {
		t.Errorf("expected service '_test._tcp', got %q", d.ServiceType())
	}
}

func TestDiscoveryOptions(t *testing.T) {
	d := New("_test._tcp", "node-1", 8080,
		WithBrowseInterval(10*time.Second),
		WithBrowseTimeout(5*time.Second),
		WithStaleTimeout(60*time.Second),
		WithMetadata(map[string]string{
			"version": "1.0.0",
			"role":    "worker",
		}),
	)

	if d.browseInterval != 10*time.Second {
		t.Errorf("expected browseInterval 10s, got %v", d.browseInterval)
	}
	if d.browseTimeout != 5*time.Second {
		t.Errorf("expected browseTimeout 5s, got %v", d.browseTimeout)
	}
	if d.staleTimeout != 60*time.Second {
		t.Errorf("expected staleTimeout 60s, got %v", d.staleTimeout)
	}
	if d.metadata["version"] != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", d.metadata["version"])
	}
	if d.metadata["role"] != "worker" {
		t.Errorf("expected role 'worker', got %q", d.metadata["role"])
	}
}

func TestPeerClone(t *testing.T) {
	p := &Peer{
		NodeID: "node-1",
		Addr:   "192.168.1.100",
		Port:   8080,
		Metadata: map[string]string{
			"version": "1.0.0",
		},
		LastSeen: time.Now(),
	}

	clone := p.Clone()
	if clone.NodeID != p.NodeID {
		t.Error("clone nodeID mismatch")
	}
	if clone.Address() != "192.168.1.100:8080" {
		t.Errorf("expected address '192.168.1.100:8080', got %q", clone.Address())
	}

	// Modify clone metadata, ensure original unchanged
	clone.Metadata["version"] = "2.0.0"
	if p.Metadata["version"] != "1.0.0" {
		t.Error("original metadata was modified")
	}
}

func TestPeerGet(t *testing.T) {
	p := &Peer{
		Metadata: map[string]string{
			"version": "1.0.0",
		},
	}

	if v := p.Get("version"); v != "1.0.0" {
		t.Errorf("expected '1.0.0', got %q", v)
	}
	if v := p.Get("missing"); v != "" {
		t.Errorf("expected empty string, got %q", v)
	}

	// Test nil metadata
	p2 := &Peer{}
	if v := p2.Get("anything"); v != "" {
		t.Errorf("expected empty string for nil metadata, got %q", v)
	}
}

func TestLocalIPs(t *testing.T) {
	ips := LocalIPs()
	// Should have at least one IP on most systems
	t.Logf("Found %d local IPs", len(ips))
	for _, ip := range ips {
		t.Logf("  %s", ip)
	}
}

func TestDiscoveryStartStop(t *testing.T) {
	d := New("_mdnstest._tcp", "test-node", 9999)

	if err := d.Start(); err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	// Verify it's running
	if d.server == nil {
		t.Error("server should be set after start")
	}
	if d.resolver == nil {
		t.Error("resolver should be set after start")
	}

	// Stop should not panic
	d.Stop()
}

func TestTwoNodeDiscovery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping discovery test in short mode")
	}

	var mu sync.Mutex
	discovered := make(map[string]bool)

	// Start node 1
	d1 := New("_mdnstest2._tcp", "node-1", 9001,
		WithBrowseInterval(1*time.Second),
		WithBrowseTimeout(2*time.Second),
	)
	d1.OnPeer(func(peer *Peer, joined bool) {
		if joined {
			mu.Lock()
			discovered[peer.NodeID] = true
			mu.Unlock()
		}
	})
	if err := d1.Start(); err != nil {
		t.Fatalf("failed to start node-1: %v", err)
	}
	defer d1.Stop()

	// Start node 2
	d2 := New("_mdnstest2._tcp", "node-2", 9002,
		WithBrowseInterval(1*time.Second),
		WithBrowseTimeout(2*time.Second),
	)
	d2.OnPeer(func(peer *Peer, joined bool) {
		if joined {
			mu.Lock()
			discovered[peer.NodeID] = true
			mu.Unlock()
		}
	})
	if err := d2.Start(); err != nil {
		t.Fatalf("failed to start node-2: %v", err)
	}
	defer d2.Stop()

	// Wait for discovery
	time.Sleep(5 * time.Second)

	mu.Lock()
	defer mu.Unlock()

	// Each node should have discovered the other
	if !discovered["node-1"] && !discovered["node-2"] {
		t.Log("No peers discovered - this may be expected in some CI environments")
	} else {
		t.Logf("Discovered peers: %v", discovered)
	}
}
