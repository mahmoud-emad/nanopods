package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	// DefaultRegistryURL is the default URL for Docker Hub's registry.
	DefaultRegistryURL = "https://registry-1.docker.io"

	// DefaultRegistryService is the default service name for Docker Hub.
	DefaultRegistryService = "registry.docker.io"

	// manifestAccept lists the manifest media types the registry may answer with.
	manifestAccept = "application/vnd.oci.image.index.v1+json, " +
		"application/vnd.oci.image.manifest.v1+json, " +
		"application/vnd.docker.distribution.manifest.list.v2+json, " +
		"application/vnd.docker.distribution.manifest.v2+json"
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

// get performs an authenticated GET request and returns the response when the registry answers with 200 OK.
// The caller must close the response body. The accept header is skipped when empty.
func (r *Registry) get(url, repository, accept string) (*http.Response, error) {
	token, err := r.obtainAccessToken(repository, "pull")
	if err != nil {
		return nil, fmt.Errorf("obtaining access token: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.BearerToken())
	if accept != "" {
		req.Header.Set("Accept", accept)
	}

	slog.Info("requesting registry", "url", url)
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}

	slog.Info("received registry response", "status", resp.Status, "content_type", resp.Header.Get("Content-Type"))
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("registry returned status: %s", resp.Status)
	}

	return resp, nil
}

// getManifest fetches a manifest by tag or digest and decodes it into v.
func (r *Registry) getManifest(repository, reference string, v any) error {
	url := fmt.Sprintf("%s/v2/%s/manifests/%s", r.RemoteURL, repository, reference)

	resp, err := r.get(url, repository, manifestAccept)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("decoding manifest: %w", err)
	}

	return nil
}

// GetImage fetches the image index for the specified repository and tag.
func (r *Registry) GetImage(repository, tag string) (*ImageIndex, error) {
	repository = normalizeRepository(repository)

	var index ImageIndex
	if err := r.getManifest(repository, tag, &index); err != nil {
		return nil, fmt.Errorf("getting image index: %w", err)
	}

	slog.Info("fetched image index", "repository", repository, "tag", tag, "manifests", len(index.Manifests))
	return &index, nil
}

// GetManifest fetches the platform-specific image manifest using its digest.
func (r *Registry) GetManifest(repository, digest string) (*ImageManifest, error) {
	repository = normalizeRepository(repository)

	var manifest ImageManifest
	if err := r.getManifest(repository, digest, &manifest); err != nil {
		return nil, fmt.Errorf("getting image manifest: %w", err)
	}

	slog.Info("fetched image manifest", "repository", repository, "digest", digest, "layers", len(manifest.Layers))
	return &manifest, nil
}

// DownloadBlob streams a blob to dest while hashing it, and keeps the file only if its content matches the digest.
// The blob is written to a temporary file next to dest, so a failed or corrupted download never leaves a partial file at dest.
// It returns the number of bytes written.
func (r *Registry) DownloadBlob(repository, digest, dest string) (int64, error) {
	repository = normalizeRepository(repository)
	url := fmt.Sprintf("%s/v2/%s/blobs/%s", r.RemoteURL, repository, digest)

	resp, err := r.get(url, repository, "")
	if err != nil {
		return 0, fmt.Errorf("downloading blob: %w", err)
	}
	defer resp.Body.Close()

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return 0, fmt.Errorf("creating directory: %w", err)
	}

	file, err := os.CreateTemp(filepath.Dir(dest), ".blob-*")
	if err != nil {
		return 0, fmt.Errorf("creating temporary file: %w", err)
	}
	// Remove the temporary file on every failure path. After a successful rename it no longer exists.
	defer os.Remove(file.Name())

	hash := sha256.New()
	written, err := io.Copy(io.MultiWriter(file, hash), resp.Body)
	if err != nil {
		file.Close()
		return 0, fmt.Errorf("streaming blob: %w", err)
	}

	// CreateTemp creates files readable by the owner only, blobs should match the other stored files.
	if err := file.Chmod(0o644); err != nil {
		file.Close()
		return 0, fmt.Errorf("setting file mode: %w", err)
	}

	if err := file.Close(); err != nil {
		return 0, fmt.Errorf("closing temporary file: %w", err)
	}

	// Verify the downloaded content against its descriptor digest.
	actual := "sha256:" + hex.EncodeToString(hash.Sum(nil))
	if actual != digest {
		return 0, fmt.Errorf("blob digest mismatch: expected %s, got %s", digest, actual)
	}

	slog.Info("verified blob digest", "expected", digest, "actual", actual)

	if err := os.Rename(file.Name(), dest); err != nil {
		return 0, fmt.Errorf("moving blob to %s: %w", dest, err)
	}

	return written, nil
}
