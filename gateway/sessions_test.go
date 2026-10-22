// ABOUTME: Tests for the per-learner session store.
// Covers network state defaults, container registration, and blocked-request counting.
package main

import "testing"

func TestUnknownSessionHasNetworkOn(t *testing.T) {
	store := NewStore()

	status := store.Status("never-seen")

	if !status.NetworkOn {
		t.Fatal("unknown session should default to network on")
	}
	if status.HasRun {
		t.Fatal("unknown session should not report a run")
	}
}

func TestSetNetworkIsPerSession(t *testing.T) {
	store := NewStore()

	store.SetNetwork("learner-a", false)

	if store.Status("learner-a").NetworkOn {
		t.Fatal("learner-a should be off")
	}
	if !store.Status("learner-b").NetworkOn {
		t.Fatal("learner-b should be unaffected")
	}
}

func TestRegisterMapsContainerToSession(t *testing.T) {
	store := NewStore()
	store.Register("learner-a", "10.0.0.5")
	store.SetNetwork("learner-a", false)

	allowed := store.AllowEgress("10.0.0.5")

	if allowed {
		t.Fatal("container of a session with network off should be blocked")
	}
	if got := store.Status("learner-a").Blocked; got != 1 {
		t.Fatalf("blocked count = %d, want 1", got)
	}
	if ip, ok := store.ContainerIP("learner-a"); !ok || ip != "10.0.0.5" {
		t.Fatalf("ContainerIP = %q, %v; want 10.0.0.5, true", ip, ok)
	}
}

func TestUnregisteredContainerIsAllowed(t *testing.T) {
	store := NewStore()

	if !store.AllowEgress("10.0.0.9") {
		t.Fatal("containers without a session should be allowed")
	}
}

func TestReusedIPMovesToNewSession(t *testing.T) {
	store := NewStore()
	store.Register("learner-a", "10.0.0.5")
	store.SetNetwork("learner-a", false)

	store.Register("learner-b", "10.0.0.5")

	if !store.AllowEgress("10.0.0.5") {
		t.Fatal("IP now belongs to learner-b, whose network is on")
	}
}

func TestTurningNetworkOnResetsBlockedCount(t *testing.T) {
	store := NewStore()
	store.Register("learner-a", "10.0.0.5")
	store.SetNetwork("learner-a", false)
	store.AllowEgress("10.0.0.5")

	store.SetNetwork("learner-a", true)
	store.SetNetwork("learner-a", false)

	if got := store.Status("learner-a").Blocked; got != 0 {
		t.Fatalf("blocked count = %d, want 0 after a fresh outage", got)
	}
}
