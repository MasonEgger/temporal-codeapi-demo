// ABOUTME: Carries a session's Temporal database from a superseded sandbox container to its replacement,
// ABOUTME: so every run a learner makes in one session shows up in the same Web UI. Held in memory, time-limited.
package main

import (
	"sync"
	"time"
)

// Handoffs stores at most one database per session, uploaded by the container that is shutting down,
// and hands it to the container that replaced it.
type Handoffs struct {
	mu       sync.Mutex
	entries  map[string]*handoffEntry
	expected map[string]string // session -> IP of the container whose database the newest container waits for
	ttl      time.Duration
	maxBytes int64
	now      func() time.Time
}

type handoffEntry struct {
	from string
	data []byte
	at   time.Time
}

// NewHandoffs keeps each uploaded database for ttl and refuses uploads larger than maxBytes.
func NewHandoffs(ttl time.Duration, maxBytes int64, now func() time.Time) *Handoffs {
	return &Handoffs{entries: map[string]*handoffEntry{}, expected: map[string]string{}, ttl: ttl, maxBytes: maxBytes, now: now}
}

// MaxBytes is the largest upload accepted.
func (h *Handoffs) MaxBytes() int64 { return h.maxBytes }

// Expect records that the session's newest container is waiting for the database of the container at fromIP.
func (h *Handoffs) Expect(session, fromIP string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sweep()
	h.expected[session] = fromIP
}

// Put stores the database a container at fromIP uploaded for its session, replacing any earlier one.
func (h *Handoffs) Put(session, fromIP string, data []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sweep()
	h.entries[session] = &handoffEntry{from: fromIP, data: data, at: h.now()}
}

// Take returns the session's database if it came from the container the newest one is waiting for.
// A database from an older container is not returned: the waiting container keeps polling for the right one.
func (h *Handoffs) Take(session string) ([]byte, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	entry, ok := h.entries[session]
	if !ok || h.now().Sub(entry.at) > h.ttl {
		return nil, false
	}
	if entry.from != h.expected[session] {
		return nil, false
	}
	return entry.data, true
}

// sweep drops expired entries. Callers must hold mu.
func (h *Handoffs) sweep() {
	for session, entry := range h.entries {
		if h.now().Sub(entry.at) > h.ttl {
			delete(h.entries, session)
			delete(h.expected, session)
		}
	}
}
