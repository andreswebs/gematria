package gematria

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// remoteWordSource is an unexported WordSource backed by an HTTP API.
// It performs reverse lookup via GET {baseURL}/words?value=N&system=S&limit=L.
type remoteWordSource struct {
	baseURL    string
	httpClient *http.Client
	authToken  string // Bearer token; empty means unauthenticated.
}

// Compile-time assertion that *remoteWordSource satisfies WordSource.
var _ WordSource = (*remoteWordSource)(nil)

// RemoteOption is a functional option for NewRemoteWordSource.
type RemoteOption func(*remoteWordSource)

// WithAuthToken sets a Bearer token sent as "Authorization: Bearer <token>"
// on every request.
func WithAuthToken(token string) RemoteOption {
	return func(s *remoteWordSource) {
		s.authToken = token
	}
}

// WithHTTPClient replaces the default http.DefaultClient with a custom client.
func WithHTTPClient(client *http.Client) RemoteOption {
	return func(s *remoteWordSource) {
		s.httpClient = client
	}
}

// NewRemoteWordSource creates a WordSource backed by an HTTP API.
// baseURL must begin with "http://" or "https://".
// Use WithAuthToken to supply a Bearer token for authenticated endpoints.
// Use WithHTTPClient to supply a custom *http.Client.
func NewRemoteWordSource(baseURL string, opts ...RemoteOption) (WordSource, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("gematria: remote word source: baseURL must not be empty")
	}
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		return nil, fmt.Errorf("gematria: remote word source: baseURL must begin with http:// or https://, got %q", baseURL)
	}

	s := &remoteWordSource{
		baseURL:    baseURL,
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s, nil
}

// remoteResponseBody is the JSON shape of the API response.
type remoteResponseBody struct {
	Words   []remoteWordJSON `json:"words"`
	HasMore bool             `json:"hasMore"`
}

type remoteWordJSON struct {
	Hebrew          string `json:"hebrew"`
	Transliteration string `json:"transliteration"`
	Meaning         string `json:"meaning"`
}

// FindByValue queries the remote API for words whose gematria value under
// system equals value, returning at most limit results.
func (s *remoteWordSource) FindByValue(value int, system System, limit int) ([]Word, bool, error) {
	u, err := url.Parse(s.baseURL + "/words")
	if err != nil {
		return nil, false, fmt.Errorf("remote word source: %w", err)
	}
	q := u.Query()
	q.Set("value", strconv.Itoa(value))
	q.Set("system", string(system))
	q.Set("limit", strconv.Itoa(limit))
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, false, fmt.Errorf("remote word source: %w", err)
	}
	if s.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.authToken)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("remote word source: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("remote word source: HTTP %d from %s", resp.StatusCode, u.String())
	}

	// Limit response body to 1 MiB to prevent memory exhaustion.
	body := io.LimitReader(resp.Body, 1<<20)
	var result remoteResponseBody
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return nil, false, fmt.Errorf("remote word source: decode: %w", err)
	}

	words := make([]Word, 0, len(result.Words))
	for _, w := range result.Words {
		words = append(words, Word(w))
	}
	return words, result.HasMore, nil
}
