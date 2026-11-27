package registry

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

const (
	// DefaultTokenURL is the default URL for obtaining access tokens from the Docker registry.
	DefaultTokenURL = "https://auth.docker.io/token"

	// defaultTokenLifetime is used when the token service does not say how long a token lives, as the token specification suggests.
	defaultTokenLifetime = 60 * time.Second

	// tokenExpiryMargin makes a cached token expire early, so it never expires while a request is in flight.
	tokenExpiryMargin = 10 * time.Second
)

// cachedToken is a token together with the time it stops being usable.
type cachedToken struct {
	token     TokenResponse
	expiresAt time.Time
}

// TokenResponse represents the response from the Docker registry token service.
type TokenResponse struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// BearerToken returns the bearer token from the TokenResponse. It prioritizes the "token" field over "access_token".
func (t TokenResponse) BearerToken() string {
	if t.Token != "" {
		return t.Token
	}
	return t.AccessToken
}

// obtainAccessToken returns an access token for the specified repository and action.
// A token is requested once and reused until it expires.
func (r *Registry) obtainAccessToken(repository string, action string) (TokenResponse, error) {
	key := repository + ":" + action

	if cached, ok := r.tokens[key]; ok && time.Now().Before(cached.expiresAt) {
		slog.Info("reusing access token", "repository", repository, "action", action, "expires_in", int(time.Until(cached.expiresAt).Seconds()))
		return cached.token, nil
	}

	token, err := r.requestAccessToken(repository, action)
	if err != nil {
		return TokenResponse{}, err
	}

	lifetime := time.Duration(token.ExpiresIn) * time.Second
	if lifetime <= 0 {
		lifetime = defaultTokenLifetime
	}

	if r.tokens == nil {
		r.tokens = make(map[string]cachedToken)
	}
	r.tokens[key] = cachedToken{token: token, expiresAt: time.Now().Add(lifetime - tokenExpiryMargin)}

	return token, nil
}

// requestAccessToken requests a new access token for the specified repository and action from the Docker registry.
func (r *Registry) requestAccessToken(repository string, action string) (TokenResponse, error) {
	u, err := url.Parse(DefaultTokenURL)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("parsing token URL: %w", err)
	}

	q := u.Query()
	q.Set("service", DefaultRegistryService)
	q.Set("scope", fmt.Sprintf("repository:%s:%s", repository, action))
	u.RawQuery = q.Encode()

	slog.Info("requesting access token", "url", u.String())
	resp, err := http.Get(u.String())
	if err != nil {
		return TokenResponse{}, fmt.Errorf("obtaining access token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return TokenResponse{}, fmt.Errorf(
			"token request failed: %s", resp.Status,
		)
	}

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return TokenResponse{}, fmt.Errorf("decoding token response: %w", err)
	}

	if token.BearerToken() == "" {
		return TokenResponse{}, fmt.Errorf("token response contains no token")
	}

	slog.Info("obtained access token", "repository", repository, "action", action, "expires_in", token.ExpiresIn)
	return token, nil
}
