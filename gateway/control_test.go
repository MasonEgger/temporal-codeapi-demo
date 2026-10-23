// ABOUTME: Tests for the browser-facing site: session API, Web UI routing, and sandbox registration.
// Uses httptest servers in place of sandbox containers.
package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

const sessionID = "6f1c2a9e-3b7d-4c55-9e0a-1b2c3d4e5f60"

func newSite(store *Store, uiPort string) *httptest.Server {
	pages := fstest.MapFS{"index.html": {Data: []byte("tutorial home")}}
	examples := fstest.MapFS{"python/shared.py": {Data: []byte("TASK_QUEUE_NAME = 'ip-address-python'")}}
	return httptest.NewServer(NewSite(store, pages, examples, SiteOptions{UIPort: uiPort}))
}

func TestSessionStatusDefaultsToNetworkOn(t *testing.T) {
	site := newSite(NewStore(), "8233")
	defer site.Close()

	response, err := http.Get(site.URL + "/api/sessions/" + sessionID)
	if err != nil {
		t.Fatal(err)
	}
	var status SessionStatus
	json.NewDecoder(response.Body).Decode(&status)

	if !status.NetworkOn || status.HasRun {
		t.Fatalf("status = %+v, want network on and no run", status)
	}
}

func TestSetNetworkOff(t *testing.T) {
	store := NewStore()
	site := newSite(store, "8233")
	defer site.Close()

	request, _ := http.NewRequest(http.MethodPut, site.URL+"/api/sessions/"+sessionID+"/network", strings.NewReader(`{"on": false}`))
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d, want 200", response.StatusCode)
	}
	if store.Status(sessionID).NetworkOn {
		t.Fatal("network should be off")
	}
}

func TestRejectsMalformedSessionID(t *testing.T) {
	site := newSite(NewStore(), "8233")
	defer site.Close()

	response, err := http.Get(site.URL + "/api/sessions/short")
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status code = %d, want 400", response.StatusCode)
	}
}

func TestServesTutorialFiles(t *testing.T) {
	site := newSite(NewStore(), "8233")
	defer site.Close()

	response, err := http.Get(site.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)

	if string(body) != "tutorial home" {
		t.Fatalf("body = %q", body)
	}
}

func TestSiteFilesAlwaysRevalidate(t *testing.T) {
	site := newSite(NewStore(), "8233")
	defer site.Close()

	response, err := http.Get(site.URL + "/")
	if err != nil {
		t.Fatal(err)
	}

	if got := response.Header.Get("Cache-Control"); got != "no-cache" {
		t.Fatalf("Cache-Control = %q, want no-cache so edited pages show up on reload", got)
	}
}

func TestServesExampleFiles(t *testing.T) {
	site := newSite(NewStore(), "8233")
	defer site.Close()

	response, err := http.Get(site.URL + "/examples/python/shared.py")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)

	if !strings.Contains(string(body), "ip-address-python") {
		t.Fatalf("body = %q", body)
	}
}

func TestWebUIBeforeAnyRunExplainsWhatToDo(t *testing.T) {
	site := newSite(NewStore(), "8233")
	defer site.Close()

	response, err := http.Get(site.URL + "/ui/" + sessionID + "/")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)

	if response.StatusCode != http.StatusNotFound || !strings.Contains(string(body), "Run") {
		t.Fatalf("got %d %q, want a 404 telling the learner to run the program", response.StatusCode, body)
	}
}

func TestWebUIIsProxiedToSessionContainer(t *testing.T) {
	var gotPath string
	webUI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		io.WriteString(w, "temporal web ui")
	}))
	defer webUI.Close()
	uiURL, _ := url.Parse(webUI.URL)
	store := NewStore()
	store.Register(sessionID, uiURL.Hostname())
	site := newSite(store, uiURL.Port())
	defer site.Close()

	response, err := http.Get(site.URL + "/ui/" + sessionID + "/namespaces/default/workflows")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)

	if string(body) != "temporal web ui" {
		t.Fatalf("body = %q", body)
	}
	// The dev server runs with --ui-public-path /ui/<session>, so the prefix is kept.
	if gotPath != "/ui/"+sessionID+"/namespaces/default/workflows" {
		t.Fatalf("upstream path = %q", gotPath)
	}
}

func TestSandboxRegistrationUsesCallerIP(t *testing.T) {
	store := NewStore()
	proxy := httptest.NewServer(NewSandboxHandler(store, NewHandoffs(time.Minute, 1<<20, time.Now), NewForwardProxy(store, nil)))
	defer proxy.Close()

	response, err := http.Post(proxy.URL+"/register/"+sessionID, "text/plain", nil)
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d, want 200", response.StatusCode)
	}
	if ip, ok := store.ContainerIP(sessionID); !ok || ip != "127.0.0.1" {
		t.Fatalf("ContainerIP = %q, %v; want 127.0.0.1", ip, ok)
	}
}

func TestCurrentReportsLatestContainerIP(t *testing.T) {
	store := NewStore()
	store.Register(sessionID, "10.0.0.8")
	proxy := httptest.NewServer(NewSandboxHandler(store, NewHandoffs(time.Minute, 1<<20, time.Now), NewForwardProxy(store, nil)))
	defer proxy.Close()

	response, err := http.Get(proxy.URL + "/current/" + sessionID)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)

	if string(body) != "10.0.0.8" {
		t.Fatalf("body = %q, want 10.0.0.8", body)
	}
}
