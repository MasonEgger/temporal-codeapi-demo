// ABOUTME: Forwards the page's Codapi API calls (/v1/) to the Codapi server so the browser sees one origin.
// Applies the per-address run limit to /v1/exec before anything reaches Codapi.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// newCodapiHandler proxies /v1/ to codapi, refusing runs over the limit with a 429 the page can explain.
func newCodapiHandler(codapi *url.URL, limiter *Limiter, trustProxy bool) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(codapi)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("codapi: %v", err)
		http.Error(w, "The playground's run service is not responding. Try again in a moment.", http.StatusBadGateway)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if limiter != nil && r.URL.Path == "/v1/exec" {
			address := requestIP(r, trustProxy)
			if !limiter.Allow(address) {
				log.Printf("codapi: run limit reached for %s", address)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{
					"stderr": "Too many runs from your address in a short time. Wait a minute and try again.",
				})
				return
			}
		}
		proxy.ServeHTTP(w, r)
	})
}

// requestIP is the address a run is attributed to: the first X-Forwarded-For entry when a trusted
// reverse proxy (Caddy) sits in front, otherwise the socket's peer address.
func requestIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			first, _, _ := strings.Cut(forwarded, ",")
			return strings.TrimSpace(first)
		}
	}
	return clientIP(r)
}
