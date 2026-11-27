package image

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/mahmoud-emad/nanopods/internal/registry"
)

const (
	// storeDirName is the directory under the user's home that holds nanopods data.
	storeDirName = ".nanopods"

	// manifestFile and configFile are the names of the image metadata files.
	manifestFile = "manifest.json"
	configFile   = "config.json"

	// layersDir is the directory holding the layer blobs of an image.
	layersDir = "layers"
)

// Dir returns the directory where the image is stored: ~/.nanopods/images/<name>/<tag>.
func (img *Image) Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding home directory: %w", err)
	}

	return filepath.Join(home, storeDirName, "images", filepath.FromSlash(img.Name), img.Tag), nil
}

// save writes data to the given path relative to the image directory, creating directories as needed.
func (img *Image) save(path string, data []byte) error {
	dir, err := img.Dir()
	if err != nil {
		return err
	}

	full := filepath.Join(dir, path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	if err := os.WriteFile(full, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", full, err)
	}

	slog.Info("stored file", "path", full, "bytes", len(data))
	return nil
}

// layerPath returns the path of a layer blob relative to the image directory.
// The digest algorithm prefix is replaced so the name is safe on every filesystem.
func layerPath(digest string) string {
	return filepath.Join(layersDir, strings.Replace(digest, ":", "-", 1))
}

// exists reports whether a file at the given path relative to the image directory exists.
func (img *Image) exists(path string) bool {
	dir, err := img.Dir()
	if err != nil {
		return false
	}

	_, err = os.Stat(filepath.Join(dir, path))
	return err == nil
}

// IsStored reports whether the image was completely downloaded before.
// It checks that the manifest, the config and every layer listed in the manifest are on disk.
func (img *Image) IsStored() bool {
	dir, err := img.Dir()
	if err != nil {
		return false
	}

	data, err := os.ReadFile(filepath.Join(dir, manifestFile))
	if err != nil {
		return false
	}

	var manifest registry.ImageManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		slog.Warn("ignoring unreadable stored manifest", "name", img.Name, "tag", img.Tag, "error", err)
		return false
	}

	if !img.exists(configFile) {
		return false
	}

	for _, layer := range manifest.Layers {
		if !img.exists(layerPath(layer.Digest)) {
			return false
		}
	}

	return true
}
