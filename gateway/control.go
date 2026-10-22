// ABOUTME: HTTP handlers for the gateway's two faces: the browser-facing site and the sandbox-facing port.
// The site serves tutorial pages, the network-switch API, and each session's Web UI; the sandbox port adds registration.
package main

import (
	"encoding/json"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
)

// sessionIDPattern accepts the random UUIDs the tutorial page generates.
var sessionIDPattern = regexp.MustCompile(`^[A-Za-z0-9-]{16,64}$`)

// SiteOptions configures the browser-facing handler.
type SiteOptions struct {
	UIPort     string   // Web UI port inside each sandbox container
	Codapi     *url.URL // Codapi server to forward /v1/ to; nil leaves the route out
	RunLimiter *Limiter // per-address cap on /v1/exec; nil means unlimited
	TrustProxy bool     // key the run limit on X-Forwarded-For (set when Caddy is in front)
}

// NewSite returns the browser-facing handler: tutorial pages at /, example sources at /examples/,
// the session API, each session's Web UI, and, when configured, the Codapi API at /v1/.
func NewSite(store *Store, pages fs.FS, examples fs.FS, options SiteOptions) http.Handler {
	uiPort := options.UIPort
	mux := http.NewServeMux()
	mux.Handle("/", revalidate(http.FileServerFS(pages)))
	mux.Handle("/examples/", revalidate(http.StripPrefix("/examples", http.FileServerFS(examples))))
	if options.Codapi != nil {
		mux.Handle("/v1/", newCodapiHandler(options.Codapi, options.RunLimiter, options.TrustProxy))
	}

	mux.HandleFunc("GET /api/sessions/{id}", withSessionID(func(w http.ResponseWriter, r *http.Request, id string) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(store.Status(id))
	}))

	mux.HandleFunc("PUT /api/sessions/{id}/network", withSessionID(func(w http.ResponseWriter, r *http.Request, id string) {
		var body struct {
			On bool `json:"on"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "body must be {\"on\": true|false}", http.StatusBadRequest)
			return
		}
		store.SetNetwork(id, body.On)
		log.Printf("site: session %s network on=%t", id, body.On)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(store.Status(id))
	}))

	mux.HandleFunc("/ui/{id}/", withSessionID(func(w http.ResponseWriter, r *http.Request, id string) {
		containerIP, ok := store.ContainerIP(id)
		if !ok {
			http.Error(w, "No Temporal Service yet. Run the program first, then reload this panel.", http.StatusNotFound)
			return
		}
		// The dev server serves the UI under /ui/<session>, so the path passes through unchanged.
		target := &url.URL{Scheme: "http", Host: net.JoinHostPort(containerIP, uiPort)}
		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, "This run's Temporal Service has shut down. Run the program again to start a new one.", http.StatusBadGateway)
		}
		proxy.ServeHTTP(w, r)
	}))

	return mux
}

// NewSandboxHandler returns the handler sandboxes reach: registration, the current-container
// lookup that lets superseded dev servers exit early, and the forward proxy.
func NewSandboxHandler(store *Store, proxy *ForwardProxy) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.IsAbs() || r.Method == http.MethodConnect {
			proxy.ServeHTTP(w, r)
			return
		}

		if id, ok := strings.CutPrefix(r.URL.Path, "/register/"); ok && r.Method == http.MethodPost {
			if !sessionIDPattern.MatchString(id) {
				http.Error(w, "invalid session id", http.StatusBadRequest)
				return
			}
			store.Register(id, clientIP(r))
			log.Printf("sandbox: session %s registered container %s", id, clientIP(r))
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if id, ok := strings.CutPrefix(r.URL.Path, "/current/"); ok && r.Method == http.MethodGet {
			containerIP, _ := store.ContainerIP(id)
			w.Write([]byte(containerIP))
			return
		}

		http.NotFound(w, r)
	})
}

// revalidate makes browsers check for a newer copy of each page file on every load,
// so edits to tutorials and examples show up on a normal reload.
func revalidate(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		handler.ServeHTTP(w, r)
	})
}

// withSessionID validates the {id} path value before calling handler.
func withSessionID(handler func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if !sessionIDPattern.MatchString(id) {
			http.Error(w, "invalid session id", http.StatusBadRequest)
			return
		}
		handler(w, r, id)
	}
}
