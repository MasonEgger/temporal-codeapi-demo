// ABOUTME: Tests for the site's /v1/ route, which forwards Codapi API calls so the page has one origin.
// Covers forwarding, the per-address run limit, and which address the limit keys on behind Caddy.
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

func fakeCodapi(t *testing.T) (*httptest.Server, *http.Request, *[]byte) {
	t.Helper()
	got := &http.Request{}
	var body []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*got = *r
		body, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"ok":true,"duration":1,"stdout":"hi","stderr":""}`)
	}))
	t.Cleanup(server.Close)
	return server, got, &body
}

func newSiteWithCodapi(t *testing.T, codapi *httptest.Server, options SiteOptions) *httptest.Server {
	t.Helper()
	pages := fstest.MapFS{"index.html": {Data: []byte("tutorial home")}}
	codapiURL, _ := url.Parse(codapi.URL)
	options.UIPort = "8233"
	options.Codapi = codapiURL
	server := httptest.NewServer(NewSite(NewStore(), pages, fstest.MapFS{}, options))
	t.Cleanup(server.Close)
	return server
}

func TestExecIsForwardedToCodapi(t *testing.T) {
	codapi, got, body := fakeCodapi(t)
	site := newSiteWithCodapi(t, codapi, SiteOptions{})

	response, err := http.Post(site.URL+"/v1/exec", "application/json", strings.NewReader(`{"sandbox":"temporal-python"}`))
	if err != nil {
		t.Fatal(err)
	}
	result, _ := io.ReadAll(response.Body)

	if got.URL.Path != "/v1/exec" || got.Method != http.MethodPost {
		t.Fatalf("upstream saw %s %s, want POST /v1/exec", got.Method, got.URL.Path)
	}
	if string(*body) != `{"sandbox":"temporal-python"}` {
		t.Fatalf("upstream body = %q", *body)
	}
	if !strings.Contains(string(result), `"stdout":"hi"`) {
		t.Fatalf("response = %q, want Codapi's result passed through", result)
	}
}

func TestExecIsNotCached(t *testing.T) {
	codapi, _, _ := fakeCodapi(t)
	site := newSiteWithCodapi(t, codapi, SiteOptions{})

	response, err := http.Post(site.URL+"/v1/exec", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}

	if got := response.Header.Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
}

func TestWithoutCodapiTheRouteIsAbsent(t *testing.T) {
	site := newSite(NewStore(), "8233")
	defer site.Close()

	response, err := http.Post(site.URL+"/v1/exec", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 when no Codapi address is configured", response.StatusCode)
	}
}

func TestRunsOverTheLimitGet429WithAMessage(t *testing.T) {
	codapi, _, _ := fakeCodapi(t)
	now := time.Unix(0, 0)
	limiter := NewLimiter(1, 1, func() time.Time { return now })
	site := newSiteWithCodapi(t, codapi, SiteOptions{RunLimiter: limiter})

	first, _ := http.Post(site.URL+"/v1/exec", "application/json", strings.NewReader(`{}`))
	second, err := http.Post(site.URL+"/v1/exec", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	var body struct {
		Stderr string `json:"stderr"`
	}
	json.NewDecoder(second.Body).Decode(&body)

	if first.StatusCode != http.StatusOK {
		t.Fatalf("first run status = %d, want 200", first.StatusCode)
	}
	if second.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("second run status = %d, want 429", second.StatusCode)
	}
	if !strings.Contains(body.Stderr, "address") {
		t.Fatalf("stderr = %q, want a message blaming this address, not the pool", body.Stderr)
	}
}

func TestLimitOnlyAppliesToExec(t *testing.T) {
	codapi, _, _ := fakeCodapi(t)
	limiter := NewLimiter(1, 1, time.Now)
	site := newSiteWithCodapi(t, codapi, SiteOptions{RunLimiter: limiter})

	http.Post(site.URL+"/v1/exec", "application/json", strings.NewReader(`{}`))
	response, err := http.Get(site.URL + "/")
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusOK {
		t.Fatalf("page status = %d, want pages unaffected by the run limit", response.StatusCode)
	}
}

func TestLimitKeysOnForwardedAddressOnlyWhenTrusted(t *testing.T) {
	post := func(site *httptest.Server, forwardedFor string) int {
		request, _ := http.NewRequest(http.MethodPost, site.URL+"/v1/exec", strings.NewReader(`{}`))
		request.Header.Set("X-Forwarded-For", forwardedFor)
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		return response.StatusCode
	}

	codapi, _, _ := fakeCodapi(t)
	trusting := newSiteWithCodapi(t, codapi, SiteOptions{RunLimiter: NewLimiter(1, 1, time.Now), TrustProxy: true})
	post(trusting, "203.0.113.1, 10.0.0.9")
	if status := post(trusting, "203.0.113.2"); status != http.StatusOK {
		t.Fatalf("behind a trusted proxy, a different forwarded address got %d, want its own bucket", status)
	}

	untrusting := newSiteWithCodapi(t, codapi, SiteOptions{RunLimiter: NewLimiter(1, 1, time.Now)})
	post(untrusting, "203.0.113.1")
	if status := post(untrusting, "203.0.113.2"); status != http.StatusTooManyRequests {
		t.Fatalf("without a trusted proxy, the header was honored (got %d); the limit must key on the socket address", status)
	}
}
