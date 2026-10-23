// ABOUTME: Entry point for the gateway that sits between learners, sandboxes, and the internet.
// Serves the tutorial site on one port and the sandbox-facing proxy and registration on another.
package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	siteAddr := flag.String("site-addr", ":3129", "address for the browser-facing site, network-switch API, and Web UI")
	sandboxAddr := flag.String("sandbox-addr", ":3128", "address for the sandbox-facing forward proxy and registration")
	siteDir := flag.String("site-dir", "/srv/site", "directory of tutorial pages to serve")
	examplesDir := flag.String("examples-dir", "/srv/examples", "directory of example sources the pages load")
	uiPort := flag.String("ui-port", "8233", "Web UI port inside each sandbox container")
	allow := flag.String("allow", "icanhazip.com:443,icanhazip.com:80,ip-api.com:80", "comma-separated host:port pairs sandboxes may reach")
	codapiAddr := flag.String("codapi-addr", "", "host:port of the Codapi server to forward /v1/ to (empty disables the route)")
	runRate := flag.Int("run-rate", 0, "runs per minute one client address may start through /v1/exec (0 disables the limit)")
	runBurst := flag.Int("run-burst", 0, "runs one client address may start at once before the rate applies")
	trustProxy := flag.Bool("trust-proxy", false, "attribute runs to the first X-Forwarded-For address (only behind a reverse proxy)")
	flag.Parse()

	store := NewStore()
	allowed := strings.Split(*allow, ",")
	// A session's Temporal database follows it from one run's container to the next, so the Web UI
	// shows every run; 32 MiB is far above a session's worth of runs and 15 minutes outlives the linger.
	handoffs := NewHandoffs(15*time.Minute, 32<<20, time.Now)
	sandbox := NewSandboxHandler(store, handoffs, NewForwardProxy(store, allowed))
	options := SiteOptions{UIPort: *uiPort, TrustProxy: *trustProxy}
	if *codapiAddr != "" {
		options.Codapi = &url.URL{Scheme: "http", Host: *codapiAddr}
		log.Printf("gateway: forwarding /v1/ to %s", options.Codapi)
	}
	if *runRate > 0 {
		burst := max(*runBurst, 1)
		options.RunLimiter = NewLimiter(*runRate, burst, time.Now)
		log.Printf("gateway: limiting runs to %d/min with a burst of %d per address", *runRate, burst)
	}
	site := NewSite(store, os.DirFS(*siteDir), os.DirFS(*examplesDir), options)

	log.Printf("gateway: sandbox proxy on %s allowing %v", *sandboxAddr, allowed)
	go func() {
		log.Fatal(http.ListenAndServe(*sandboxAddr, sandbox))
	}()
	log.Printf("gateway: site on %s serving %s", *siteAddr, *siteDir)
	log.Fatal(http.ListenAndServe(*siteAddr, site))
}
