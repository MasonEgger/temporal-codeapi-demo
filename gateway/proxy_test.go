// ABOUTME: Tests for the allowlisting forward proxy that sandboxes use to reach the internet.
// Exercises plain HTTP forwarding, HTTPS tunneling, the allowlist, and the per-session network switch.
package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// newProxyClient returns an HTTP client that sends every request through proxyURL.
func newProxyClient(proxyURL string, base *http.Client) *http.Client {
	parsed, _ := url.Parse(proxyURL)
	transport := &http.Transport{Proxy: http.ProxyURL(parsed)}
	if base != nil {
		transport.TLSClientConfig = base.Transport.(*http.Transport).TLSClientConfig
	}
	return &http.Client{Transport: transport}
}

func hostPort(rawURL string) string {
	parsed, _ := url.Parse(rawURL)
	return parsed.Host
}

func TestForwardsAllowedHTTPRequest(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "203.0.113.7")
	}))
	defer upstream.Close()
	proxy := httptest.NewServer(NewForwardProxy(NewStore(), []string{hostPort(upstream.URL)}))
	defer proxy.Close()

	response, err := newProxyClient(proxy.URL, nil).Get(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)

	if response.StatusCode != http.StatusOK || string(body) != "203.0.113.7" {
		t.Fatalf("got %d %q, want 200 203.0.113.7", response.StatusCode, body)
	}
}

func TestTunnelsAllowedHTTPSRequest(t *testing.T) {
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "secure")
	}))
	defer upstream.Close()
	proxy := httptest.NewServer(NewForwardProxy(NewStore(), []string{hostPort(upstream.URL)}))
	defer proxy.Close()

	response, err := newProxyClient(proxy.URL, upstream.Client()).Get(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)

	if string(body) != "secure" {
		t.Fatalf("got %q, want secure", body)
	}
}

func TestRejectsHostOutsideAllowlist(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not reach a host outside the allowlist")
	}))
	defer upstream.Close()
	proxy := httptest.NewServer(NewForwardProxy(NewStore(), []string{"icanhazip.com:443"}))
	defer proxy.Close()

	response, err := newProxyClient(proxy.URL, nil).Get(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.StatusCode)
	}
}

func TestRejectsTunnelOutsideAllowlist(t *testing.T) {
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("tunnel should not reach a host outside the allowlist")
	}))
	defer upstream.Close()
	proxy := httptest.NewServer(NewForwardProxy(NewStore(), []string{"icanhazip.com:443"}))
	defer proxy.Close()

	_, err := newProxyClient(proxy.URL, upstream.Client()).Get(upstream.URL)

	if err == nil || !strings.Contains(err.Error(), "Forbidden") {
		t.Fatalf("err = %v, want a Forbidden CONNECT failure", err)
	}
}

func TestBlocksRequestsWhileSessionNetworkIsOff(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request should not reach upstream during an outage")
	}))
	defer upstream.Close()
	store := NewStore()
	store.Register("learner-a", "127.0.0.1")
	store.SetNetwork("learner-a", false)
	proxy := httptest.NewServer(NewForwardProxy(store, []string{hostPort(upstream.URL)}))
	defer proxy.Close()

	response, err := newProxyClient(proxy.URL, nil).Get(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", response.StatusCode)
	}
	if got := store.Status("learner-a").Blocked; got != 1 {
		t.Fatalf("blocked = %d, want 1", got)
	}
}
