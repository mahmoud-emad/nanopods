package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

const (
	// DefaultRegistryURL is the default URL for Docker Hub's registry.
	DefaultRegistryURL = "https://registry-1.docker.io"

	// DefaultRegistryService is the default service name for Docker Hub.
	DefaultRegistryService = "registry.docker.io"
)

// Registry represents a Docker registry.
type Registry struct {
	RemoteURL string
	Client    *http.Client
}

// New creates a new Registry instance.
func New() *Registry {
	return &Registry{
		RemoteURL: DefaultRegistryURL,
		Client:    http.DefaultClient,
	}
}

// normalizeRepository adds the library namespace to official Docker Hub images.
func normalizeRepository(repository string) string {
	if !strings.Contains(repository, "/") {
		return "library/" + repository
	}

	return repository
}

// getManifestResponse performs an authenticated request to fetch a manifest.
func (r *Registry) getManifestResponse(url, repository string) (*http.Response, error) {
	token, err := r.obtainAccessToken(repository, "pull")
	if err != nil {
		return nil, fmt.Errorf("obtaining access token: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
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

	slog.Info("requesting manifest", "url", url)
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching image manifest: %w", err)
	}

	slog.Info("received manifest response", "status", resp.Status, "content_type", resp.Header.Get("Content-Type"))
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("registry returned status: %s", resp.Status)
	}

	return resp, nil
}

// GetImage fetches the image index for the specified repository and tag.
func (r *Registry) GetImage(repository, tag string) (*ImageIndex, error) {
	repository = normalizeRepository(repository)

	path := fmt.Sprintf(
		"%s/v2/%s/manifests/%s",
		r.RemoteURL,
		repository,
		tag,
	)

	resp, err := r.getManifestResponse(path, repository)
	if err != nil {
		return nil, fmt.Errorf("getting image index: %w", err)
	}
	defer resp.Body.Close()

	var index ImageIndex

	if err := json.NewDecoder(resp.Body).Decode(&index); err != nil {
		return nil, fmt.Errorf("decoding image index: %w", err)
	}

	slog.Info("fetched image index", "repository", repository, "tag", tag, "manifests", len(index.Manifests))
	return &index, nil
}

// GetManifest fetches the platform-specific image manifest using its digest.
func (r *Registry) GetManifest(repository, digest string) (*ImageManifest, error) {
	repository = normalizeRepository(repository)

	path := fmt.Sprintf(
		"%s/v2/%s/manifests/%s",
		r.RemoteURL,
		repository,
		digest,
	)

	resp, err := r.getManifestResponse(path, repository)
	if err != nil {
		return nil, fmt.Errorf("getting image manifest: %w", err)
	}
	defer resp.Body.Close()

	var manifest ImageManifest

	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("decoding image manifest: %w", err)
	}

	slog.Info("fetched image manifest", "repository", repository, "digest", digest, "layers", len(manifest.Layers))
	return &manifest, nil
}

func (r *Registry) GetBlob(repository, digest string) ([]byte, error) {
	repository = normalizeRepository(repository)

	url := fmt.Sprintf(
		"%s/v2/%s/blobs/%s",
		r.RemoteURL,
		repository,
		digest,
	)

	token, err := r.obtainAccessToken(repository, "pull")
	if err != nil {
		return nil, fmt.Errorf("obtaining access token: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating blob request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.BearerToken())

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading blob: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status: %s", resp.Status)
	}

	blob, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading blob: %w", err)
	}

	// Verify the downloaded content against its descriptor digest.
	expectedDigest := strings.TrimPrefix(digest, "sha256:")
	actualDigest := sha256.Sum256(blob)
	actualHex := hex.EncodeToString(actualDigest[:])

	if actualHex != expectedDigest {
		return nil, fmt.Errorf("blob digest mismatch: expected %s, got sha256:%s",
			digest, actualHex)
	}

	return blob, nil
}
