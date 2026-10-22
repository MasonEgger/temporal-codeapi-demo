// ABOUTME: Entry point for the gateway that sits between learners, sandboxes, and the internet.
// Serves the tutorial site on one port and the sandbox-facing proxy and registration on another.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	siteAddr := flag.String("site-addr", ":3129", "address for the browser-facing site, network-switch API, and Web UI")
	sandboxAddr := flag.String("sandbox-addr", ":3128", "address for the sandbox-facing forward proxy and registration")
	siteDir := flag.String("site-dir", "/srv/site", "directory of tutorial pages to serve")
	examplesDir := flag.String("examples-dir", "/srv/examples", "directory of example sources the pages load")
	uiPort := flag.String("ui-port", "8233", "Web UI port inside each sandbox container")
	allow := flag.String("allow", "icanhazip.com:443,icanhazip.com:80,ip-api.com:80", "comma-separated host:port pairs sandboxes may reach")
	flag.Parse()

	store := NewStore()
	allowed := strings.Split(*allow, ",")
	sandbox := NewSandboxHandler(store, NewForwardProxy(store, allowed))
	site := NewSite(store, os.DirFS(*siteDir), os.DirFS(*examplesDir), *uiPort)

	log.Printf("gateway: sandbox proxy on %s allowing %v", *sandboxAddr, allowed)
	go func() {
		log.Fatal(http.ListenAndServe(*sandboxAddr, sandbox))
	}()
	log.Printf("gateway: site on %s serving %s", *siteAddr, *siteDir)
	log.Fatal(http.ListenAndServe(*siteAddr, site))
}
