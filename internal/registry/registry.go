package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	DefaultRegistryURL     = "https://registry-1.docker.io"
	DefaultRegistryService = "registry.docker.io"
	DefaultTokenURL        = "https://auth.docker.io/token"
)

type TokenResponse struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (t TokenResponse) BearerToken() string {
	if t.Token != "" {
		return t.Token
	}
	return t.AccessToken
}

// Registry interface {
//     Pull(ref Reference) (*Image, error)
// }

type Registry struct {
	RemoteURL string
}

func New() *Registry {
	return &Registry{
		RemoteURL: DefaultRegistryURL,
	}
}

func (r *Registry) obtainAccessToken(repository string, action string) (TokenResponse, error) {
	u, err := url.Parse(DefaultTokenURL)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("parsing token URL: %w", err)
	}

	q := u.Query()
	q.Set("service", DefaultRegistryService)
	q.Set("scope", fmt.Sprintf("repository:%s:%s", repository, action))
	u.RawQuery = q.Encode()

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

	return token, nil
}

func (r *Registry) GetImage(repository, tag string) (string, error) {
	path := fmt.Sprintf("%s/v2/%s/manifests/%s", r.RemoteURL, repository, tag)

	token, err := r.obtainAccessToken(repository, "pull")
	if err != nil {
		return "", fmt.Errorf("obtaining access token: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return "", fmt.Errorf("creating manifest request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.BearerToken())
	req.Header.Set("Accept",
		"application/vnd.oci.image.index.v1+json, "+
			"application/vnd.oci.image.manifest.v1+json, "+
			"application/vnd.docker.distribution.manifest.list.v2+json, "+
			"application/vnd.docker.distribution.manifest.v2+json",
	)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching image manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"failed to fetch manifest for %s:%s: %s",
			repository, tag, resp.Status,
		)
	}

	return path, nil
}
