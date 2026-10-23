// ABOUTME: Forward proxy that is the sandbox network's only route to the internet.
// Allows only listed host:port pairs and refuses all traffic from a session whose network switch is off.
package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

// hopHeaders are connection-scoped headers a proxy must not forward.
var hopHeaders = []string{
	"Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization",
	"Proxy-Connection", "Te", "Trailer", "Transfer-Encoding", "Upgrade",
}

// ForwardProxy handles plain HTTP proxy requests and HTTPS CONNECT tunnels.
type ForwardProxy struct {
	store     *Store
	allowed   map[string]bool
	transport http.RoundTripper
}

// NewForwardProxy returns a proxy that only reaches the given host:port pairs.
func NewForwardProxy(store *Store, allowed []string) *ForwardProxy {
	allowSet := make(map[string]bool, len(allowed))
	for _, target := range allowed {
		allowSet[target] = true
	}
	// The proxy itself must never use a proxy, whatever the environment says.
	transport := &http.Transport{Proxy: nil, ResponseHeaderTimeout: 15 * time.Second}
	return &ForwardProxy{store: store, allowed: allowSet, transport: transport}
}

func (p *ForwardProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var target string
	switch {
	case r.Method == http.MethodConnect:
		target = r.Host
	case r.URL.IsAbs():
		target = r.URL.Host
		if r.URL.Port() == "" {
			target = net.JoinHostPort(r.URL.Hostname(), defaultPort(r.URL.Scheme))
		}
	default:
		http.Error(w, "this is a forward proxy; send absolute URLs or CONNECT", http.StatusBadRequest)
		return
	}

	if !p.allowed[target] {
		log.Printf("proxy: denied %s (not on the allowlist)", target)
		http.Error(w, target+" is not on the sandbox allowlist", http.StatusForbidden)
		return
	}
	if !p.store.AllowEgress(clientIP(r)) {
		log.Printf("proxy: blocked %s from %s (network switched off)", target, clientIP(r))
		http.Error(w, "network is switched off for this session", http.StatusServiceUnavailable)
		return
	}

	if r.Method == http.MethodConnect {
		p.tunnel(w, target)
		return
	}
	p.forward(w, r)
}

// tunnel opens a TCP connection to target and splices it to the client connection.
func (p *ForwardProxy) tunnel(w http.ResponseWriter, target string) {
	upstream, err := net.DialTimeout("tcp", target, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	client, _, err := http.NewResponseController(w).Hijack()
	if err != nil {
		upstream.Close()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := io.WriteString(client, "HTTP/1.1 200 Connection Established\r\n\r\n"); err != nil {
		client.Close()
		upstream.Close()
		return
	}

	// Copy in both directions; closing both ends when either side finishes unblocks the other copy.
	go func() {
		io.Copy(upstream, client)
		upstream.Close()
		client.Close()
	}()
	io.Copy(client, upstream)
	upstream.Close()
	client.Close()
}

// forward relays a plain HTTP request to its absolute URL and copies back the response.
func (p *ForwardProxy) forward(w http.ResponseWriter, r *http.Request) {
	outbound := r.Clone(r.Context())
	outbound.RequestURI = ""
	for _, header := range hopHeaders {
		outbound.Header.Del(header)
	}

	response, err := p.transport.RoundTrip(outbound)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer response.Body.Close()

	for _, header := range hopHeaders {
		response.Header.Del(header)
	}
	for name, values := range response.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	w.WriteHeader(response.StatusCode)
	io.Copy(w, response.Body)
}

func defaultPort(scheme string) string {
	if scheme == "https" {
		return "443"
	}
	return "80"
}

// clientIP returns the IP address of the connection that sent r.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
