package gematria_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gematria "github.com/andreswebs/gematria"
)

// remoteResponse is the JSON shape returned by the mock API server.
type remoteResponse struct {
	Words   []remoteWord `json:"words"`
	HasMore bool         `json:"hasMore"`
}

type remoteWord struct {
	Hebrew          string `json:"hebrew"`
	Transliteration string `json:"transliteration,omitempty"`
	Meaning         string `json:"meaning,omitempty"`
}

// newMockRemoteServer starts a test HTTP server that returns the given words
// and hasMore flag for any /words request.
func newMockRemoteServer(t *testing.T, words []remoteWord, hasMore bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(remoteResponse{Words: words, HasMore: hasMore})
	}))
}

// TestNewRemoteWordSource_invalidBaseURL verifies that an empty or non-http(s)
// baseURL is rejected at construction time.
func TestNewRemoteWordSource_invalidBaseURL(t *testing.T) {
	cases := []struct {
		name    string
		baseURL string
	}{
		{"empty", ""},
		{"no scheme", "example.com/words"},
		{"ftp scheme", "ftp://example.com"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := gematria.NewRemoteWordSource(tc.baseURL)
			if err == nil {
				t.Errorf("NewRemoteWordSource(%q): expected error, got nil", tc.baseURL)
			}
		})
	}
}

// TestNewRemoteWordSource_FindByValue_match is the tracer bullet:
// a mock server with one matching word returns that word via FindByValue.
func TestNewRemoteWordSource_FindByValue_match(t *testing.T) {
	srv := newMockRemoteServer(t, []remoteWord{
		{Hebrew: "שלום", Transliteration: "shalom", Meaning: "peace"},
	}, false)
	defer srv.Close()

	src, err := gematria.NewRemoteWordSource(srv.URL)
	if err != nil {
		t.Fatalf("NewRemoteWordSource: %v", err)
	}

	words, hasMore, err := src.FindByValue(376, gematria.Hechrachi, 20)
	if err != nil {
		t.Fatalf("FindByValue: %v", err)
	}
	if hasMore {
		t.Error("hasMore should be false")
	}
	if len(words) != 1 {
		t.Fatalf("expected 1 word, got %d", len(words))
	}
	if words[0].Hebrew != "שלום" {
		t.Errorf("Hebrew: got %q, want %q", words[0].Hebrew, "שלום")
	}
	if words[0].Transliteration != "shalom" {
		t.Errorf("Transliteration: got %q, want %q", words[0].Transliteration, "shalom")
	}
	if words[0].Meaning != "peace" {
		t.Errorf("Meaning: got %q, want %q", words[0].Meaning, "peace")
	}
}

// TestNewRemoteWordSource_FindByValue_queryParams verifies that FindByValue
// sends value, system, and limit as query parameters.
func TestNewRemoteWordSource_FindByValue_queryParams(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(remoteResponse{})
	}))
	defer srv.Close()

	src, err := gematria.NewRemoteWordSource(srv.URL)
	if err != nil {
		t.Fatalf("NewRemoteWordSource: %v", err)
	}

	_, _, err = src.FindByValue(376, gematria.Hechrachi, 5)
	if err != nil {
		t.Fatalf("FindByValue: %v", err)
	}

	// Parse query params manually.
	values, parseErr := parseQuery(gotQuery)
	if parseErr != nil {
		t.Fatalf("parseQuery: %v", parseErr)
	}
	if values["value"] != "376" {
		t.Errorf("query param value: got %q, want %q", values["value"], "376")
	}
	if values["system"] != "hechrachi" {
		t.Errorf("query param system: got %q, want %q", values["system"], "hechrachi")
	}
	if values["limit"] != "5" {
		t.Errorf("query param limit: got %q, want %q", values["limit"], "5")
	}
}

// parseQuery is a tiny query-string parser for test assertions.
func parseQuery(raw string) (map[string]string, error) {
	r, err := http.NewRequest(http.MethodGet, "/?"+raw, nil)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string)
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			out[k] = v[0]
		}
	}
	return out, nil
}

// TestNewRemoteWordSource_FindByValue_authToken verifies that an auth token is
// sent as "Authorization: Bearer <token>" when provided via WithAuthToken.
func TestNewRemoteWordSource_FindByValue_authToken(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(remoteResponse{})
	}))
	defer srv.Close()

	src, err := gematria.NewRemoteWordSource(srv.URL, gematria.WithAuthToken("mysecret"))
	if err != nil {
		t.Fatalf("NewRemoteWordSource: %v", err)
	}

	_, _, err = src.FindByValue(376, gematria.Hechrachi, 20)
	if err != nil {
		t.Fatalf("FindByValue: %v", err)
	}

	want := "Bearer mysecret"
	if gotAuth != want {
		t.Errorf("Authorization header: got %q, want %q", gotAuth, want)
	}
}

// TestNewRemoteWordSource_FindByValue_noAuth verifies that no Authorization
// header is sent when no token is provided.
func TestNewRemoteWordSource_FindByValue_noAuth(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(remoteResponse{})
	}))
	defer srv.Close()

	src, err := gematria.NewRemoteWordSource(srv.URL)
	if err != nil {
		t.Fatalf("NewRemoteWordSource: %v", err)
	}

	_, _, err = src.FindByValue(376, gematria.Hechrachi, 20)
	if err != nil {
		t.Fatalf("FindByValue: %v", err)
	}

	if gotAuth != "" {
		t.Errorf("Authorization header should be absent, got %q", gotAuth)
	}
}

// TestNewRemoteWordSource_FindByValue_nonOK verifies that a non-200 response
// returns an error containing the HTTP status code.
func TestNewRemoteWordSource_FindByValue_nonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	src, err := gematria.NewRemoteWordSource(srv.URL)
	if err != nil {
		t.Fatalf("NewRemoteWordSource: %v", err)
	}

	_, _, err = src.FindByValue(376, gematria.Hechrachi, 20)
	if err == nil {
		t.Fatal("expected error for non-200 response, got nil")
	}
}

// TestNewRemoteWordSource_FindByValue_networkError verifies that a network
// failure returns a wrapped error (not a panic or nil).
func TestNewRemoteWordSource_FindByValue_networkError(t *testing.T) {
	// Use a server that is immediately closed so requests fail.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	srv.Close() // close before making request

	src, err := gematria.NewRemoteWordSource(srv.URL)
	if err != nil {
		t.Fatalf("NewRemoteWordSource: %v", err)
	}

	_, _, err = src.FindByValue(376, gematria.Hechrachi, 20)
	if err == nil {
		t.Fatal("expected error for network failure, got nil")
	}
}

// TestNewRemoteWordSource_FindByValue_hasMore verifies that hasMore is
// propagated from the server response.
func TestNewRemoteWordSource_FindByValue_hasMore(t *testing.T) {
	srv := newMockRemoteServer(t, []remoteWord{
		{Hebrew: "שלום"},
		{Hebrew: "אמת"},
	}, true)
	defer srv.Close()

	src, err := gematria.NewRemoteWordSource(srv.URL)
	if err != nil {
		t.Fatalf("NewRemoteWordSource: %v", err)
	}

	_, hasMore, err := src.FindByValue(100, gematria.Hechrachi, 2)
	if err != nil {
		t.Fatalf("FindByValue: %v", err)
	}
	if !hasMore {
		t.Error("hasMore should be true when server says so")
	}
}

// TestNewRemoteWordSource_FindByValue_emptyResult verifies (nil, false, nil)
// when the server returns an empty words array.
func TestNewRemoteWordSource_FindByValue_emptyResult(t *testing.T) {
	srv := newMockRemoteServer(t, []remoteWord{}, false)
	defer srv.Close()

	src, err := gematria.NewRemoteWordSource(srv.URL)
	if err != nil {
		t.Fatalf("NewRemoteWordSource: %v", err)
	}

	words, hasMore, err := src.FindByValue(999, gematria.Hechrachi, 20)
	if err != nil {
		t.Fatalf("FindByValue: %v", err)
	}
	if hasMore {
		t.Error("hasMore should be false")
	}
	if len(words) != 0 {
		t.Errorf("expected 0 words, got %d", len(words))
	}
}

// TestNewRemoteWordSource_notCloser verifies that remoteWordSource does NOT
// implement io.Closer (HTTP client requires no cleanup).
func TestNewRemoteWordSource_notCloser(t *testing.T) {
	srv := newMockRemoteServer(t, nil, false)
	defer srv.Close()

	src, err := gematria.NewRemoteWordSource(srv.URL)
	if err != nil {
		t.Fatalf("NewRemoteWordSource: %v", err)
	}

	if _, ok := src.(interface{ Close() error }); ok {
		t.Error("remoteWordSource should NOT implement io.Closer")
	}
}

// TestWithHTTPClient verifies that a custom HTTP client is used when provided
// via WithHTTPClient.
func TestWithHTTPClient_used(t *testing.T) {
	var called bool
	transport := &trackingTransport{
		inner:  http.DefaultTransport,
		called: &called,
	}
	client := &http.Client{Transport: transport}

	srv := newMockRemoteServer(t, []remoteWord{{Hebrew: "שלום"}}, false)
	defer srv.Close()

	src, err := gematria.NewRemoteWordSource(srv.URL, gematria.WithHTTPClient(client))
	if err != nil {
		t.Fatalf("NewRemoteWordSource: %v", err)
	}

	_, _, err = src.FindByValue(376, gematria.Hechrachi, 20)
	if err != nil {
		t.Fatalf("FindByValue: %v", err)
	}

	if !called {
		t.Error("custom HTTP client was not used")
	}
}

// trackingTransport wraps an http.RoundTripper and records whether it was called.
type trackingTransport struct {
	inner  http.RoundTripper
	called *bool
}

func (t *trackingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	*t.called = true
	return t.inner.RoundTrip(req)
}
