// ABOUTME: In-memory store of learner sessions for the gateway.
// Tracks each session's network switch, its latest sandbox container, and requests blocked during an outage.
package main

import "sync"

// SessionStatus is what the tutorial page shows about a learner's session.
type SessionStatus struct {
	NetworkOn bool `json:"networkOn"`
	Blocked   int  `json:"blocked"`
	HasRun    bool `json:"hasRun"`
}

type session struct {
	networkOff  bool
	blocked     int
	containerIP string
}

// Store maps learner sessions to their network state and sandbox containers.
type Store struct {
	mu        sync.Mutex
	sessions  map[string]*session
	ipSession map[string]string
}

// NewStore returns an empty Store.
func NewStore() *Store {
	return &Store{sessions: map[string]*session{}, ipSession: map[string]string{}}
}

// get returns the session for id, creating it if needed. Callers must hold mu.
func (s *Store) get(id string) *session {
	sess, ok := s.sessions[id]
	if !ok {
		sess = &session{}
		s.sessions[id] = sess
	}
	return sess
}

// Status reports a session's network state, blocked count, and whether it has a container.
func (s *Store) Status(id string) SessionStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[id]
	if !ok {
		return SessionStatus{NetworkOn: true}
	}
	return SessionStatus{NetworkOn: !sess.networkOff, Blocked: sess.blocked, HasRun: sess.containerIP != ""}
}

// SetNetwork flips a session's network switch. Starting an outage resets the blocked count.
func (s *Store) SetNetwork(id string, on bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess := s.get(id)
	if !on && !sess.networkOff {
		sess.blocked = 0
	}
	sess.networkOff = !on
}

// Register records that a session's latest sandbox container has the given IP.
func (s *Store) Register(id, containerIP string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.get(id).containerIP = containerIP
	s.ipSession[containerIP] = id
}

// ContainerIP returns the IP of a session's latest sandbox container.
func (s *Store) ContainerIP(id string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[id]
	if !ok || sess.containerIP == "" {
		return "", false
	}
	return sess.containerIP, true
}

// AllowEgress reports whether a container may reach the internet, counting the request if not.
func (s *Store) AllowEgress(containerIP string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.ipSession[containerIP]
	if !ok {
		return true
	}
	sess := s.sessions[id]
	if sess.networkOff {
		sess.blocked++
		return false
	}
	return true
}
