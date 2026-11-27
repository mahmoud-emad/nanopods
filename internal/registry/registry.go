package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	// DefaultRegistryURL is the default URL for the Docker registry.
	DefaultRegistryURL = "https://registry-1.docker.io"
	// DefaultRegistryService is the default service name for the Docker registry.
	DefaultRegistryService = "registry.docker.io"
)

// Registry represents a Docker registry with its remote URL.
type Registry struct {
	RemoteURL string
}

// New creates a new Registry instance with the default remote URL.
func New() *Registry {
	return &Registry{
		RemoteURL: DefaultRegistryURL,
	}
}

// GetImage fetches the image manifest from the registry for the specified repository and tag. It returns the manifest path or an error if the image is not found.
func (r *Registry) GetImage(repository, tag string) (*ImageIndex, error) {
	path := fmt.Sprintf("%s/v2/library/%s/manifests/%s", r.RemoteURL, repository, tag)

	token, err := r.obtainAccessToken("library/"+repository, "pull")
	if err != nil {
		return nil, fmt.Errorf("obtaining access token: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("creating manifest request: %w", err)
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
		return nil, fmt.Errorf("fetching image manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"failed to fetch manifest for %s:%s: %s",
			repository, tag, resp.Status,
		)
	}

	var index *ImageIndex
	if err := json.NewDecoder(resp.Body).Decode(&index); err != nil {
		return nil, fmt.Errorf("decoding manifest response: %w", err)
	}

	return index, nil

}
