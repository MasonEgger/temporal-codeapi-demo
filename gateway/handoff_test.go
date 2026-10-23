// ABOUTME: Tests for the session database handoff: which uploads a replacement container may take,
// ABOUTME: how long they live, and the sandbox-facing routes that carry them.
package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTakeReturnsOnlyTheExpectedContainersDatabase(t *testing.T) {
	now := time.Unix(0, 0)
	handoffs := NewHandoffs(time.Minute, 1<<20, func() time.Time { return now })

	handoffs.Put(sessionID, "10.0.0.8", []byte("old database"))
	if _, ok := handoffs.Take(sessionID); ok {
		t.Fatal("nothing should be taken before a replacement container registers")
	}

	handoffs.Expect(sessionID, "10.0.0.9")
	if _, ok := handoffs.Take(sessionID); ok {
		t.Fatal("a database from a container other than the superseded one is stale")
	}

	handoffs.Put(sessionID, "10.0.0.9", []byte("current database"))
	data, ok := handoffs.Take(sessionID)
	if !ok || string(data) != "current database" {
		t.Fatalf("Take = %q, %v; want the superseded container's database", data, ok)
	}
}

func TestHandoffsExpire(t *testing.T) {
	now := time.Unix(0, 0)
	handoffs := NewHandoffs(time.Minute, 1<<20, func() time.Time { return now })
	handoffs.Expect(sessionID, "10.0.0.8")
	handoffs.Put(sessionID, "10.0.0.8", []byte("database"))

	now = now.Add(2 * time.Minute)
	if _, ok := handoffs.Take(sessionID); ok {
		t.Fatal("an upload older than the ttl must not be handed over")
	}
}

func newSandbox(t *testing.T, store *Store, handoffs *Handoffs) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(NewSandboxHandler(store, handoffs, NewForwardProxy(store, nil)))
	t.Cleanup(server.Close)
	return server
}

func TestRegistrationSaysWhetherToWaitForAHandoff(t *testing.T) {
	store := NewStore()
	handoffs := NewHandoffs(time.Minute, 1<<20, time.Now)
	sandbox := newSandbox(t, store, handoffs)

	first, _ := http.Post(sandbox.URL+"/register/"+sessionID, "text/plain", nil)
	firstBody, _ := io.ReadAll(first.Body)
	if string(firstBody) != "none" {
		t.Fatalf("first registration answered %q, want none (no previous container)", firstBody)
	}

	// A later container for the same session arrives from another address.
	store.Register(sessionID, "10.0.0.8")
	second, _ := http.Post(sandbox.URL+"/register/"+sessionID, "text/plain", nil)
	secondBody, _ := io.ReadAll(second.Body)
	if string(secondBody) != "handoff" {
		t.Fatalf("second registration answered %q, want handoff (a previous container exists)", secondBody)
	}
}

func TestDatabaseRoundTripsThroughTheGateway(t *testing.T) {
	store := NewStore()
	handoffs := NewHandoffs(time.Minute, 1<<20, time.Now)
	sandbox := newSandbox(t, store, handoffs)
	// The test client is the superseded container as far as the gateway can tell.
	handoffs.Expect(sessionID, "127.0.0.1")

	before, _ := http.Get(sandbox.URL + "/handoff/" + sessionID)
	if before.StatusCode != http.StatusNotFound {
		t.Fatalf("status before upload = %d, want 404 so the new container keeps polling", before.StatusCode)
	}

	upload, err := http.Post(sandbox.URL+"/handoff/"+sessionID, "application/octet-stream", bytes.NewReader([]byte("sqlite bytes")))
	if err != nil {
		t.Fatal(err)
	}
	if upload.StatusCode != http.StatusNoContent {
		t.Fatalf("upload status = %d, want 204", upload.StatusCode)
	}

	download, _ := http.Get(sandbox.URL + "/handoff/" + sessionID)
	body, _ := io.ReadAll(download.Body)
	if download.StatusCode != http.StatusOK || string(body) != "sqlite bytes" {
		t.Fatalf("download = %d %q, want 200 and the uploaded bytes", download.StatusCode, body)
	}
}

func TestOversizedUploadIsRefused(t *testing.T) {
	store := NewStore()
	handoffs := NewHandoffs(time.Minute, 16, time.Now)
	sandbox := newSandbox(t, store, handoffs)

	upload, err := http.Post(sandbox.URL+"/handoff/"+sessionID, "application/octet-stream", bytes.NewReader(make([]byte, 17)))
	if err != nil {
		t.Fatal(err)
	}

	if upload.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", upload.StatusCode)
	}
}
