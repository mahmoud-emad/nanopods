package registry

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
)

// DefaultTokenURL is the default URL for obtaining access tokens from the Docker registry.
const DefaultTokenURL = "https://auth.docker.io/token"

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

// obtainAccessToken obtains an access token for the specified repository and action from the Docker registry.
func (r *Registry) obtainAccessToken(repository string, action string) (TokenResponse, error) {
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
