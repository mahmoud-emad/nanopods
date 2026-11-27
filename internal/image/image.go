package image

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/mahmoud-emad/nanopods/internal/registry"
	"github.com/mahmoud-emad/nanopods/internal/utils"
)

var (
	// ErrImageNotFound is returned when the requested image is not found in the registry.
	ErrImageNotFound = fmt.Errorf("image not found")
)

type Image struct {
	ID   string
	Name string
	Tag  string
}

// New creates a new Image instance with the given name and tag. It generates a random ID for the image.
func New(name, tag string) *Image {
	return &Image{
		ID:   utils.RandomString(10),
		Name: name,
		Tag:  tag,
	}
}

// Pull downloads the image manifest, config and layers from the default registry and stores them in the image directory.
func (img *Image) Pull() (Image, error) {
	// We should later change this to use a registry interface to support multiple registries, for now we will use the default registry.
	r := registry.New()

	slog.Info("pulling image", "name", img.Name, "tag", img.Tag, "id", img.ID)
	index, err := r.GetImage(img.Name, img.Tag)
	if err != nil {
		return Image{}, fmt.Errorf("error pulling image: %w", err)
	}

	// Find the manifest matching the host platform.
	// For now, let's assume linux/amd64. TODO: We need to support different platforms.
	descriptor, err := index.SelectPlatform("linux", "amd64")
	if err != nil {
		return Image{}, fmt.Errorf("cannot select platform: %w", err)
	}

	slog.Info("selected platform", "digest", descriptor.Digest, "os", descriptor.Platform.OS, "arch", descriptor.Platform.Architecture)

	manifest, err := r.GetManifest(img.Name, descriptor.Digest)
	if err != nil {
		return Image{}, fmt.Errorf("cannot fetch image manifest: %w", err)
	}

	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return Image{}, fmt.Errorf("cannot encode image manifest: %w", err)
	}
	if err := img.save(manifestFile, manifestJSON); err != nil {
		return Image{}, fmt.Errorf("cannot store image manifest: %w", err)
	}

	config, err := r.GetBlob(img.Name, manifest.Config.Digest)
	if err != nil {
		return Image{}, fmt.Errorf("cannot download config blob: %w", err)
	}
	slog.Info("downloaded config", "digest", manifest.Config.Digest, "bytes", len(config))
	if err := img.save(configFile, config); err != nil {
		return Image{}, fmt.Errorf("cannot store config blob: %w", err)
	}

	for i, layer := range manifest.Layers {
		blob, err := r.GetBlob(img.Name, layer.Digest)
		if err != nil {
			return Image{}, fmt.Errorf("cannot download layer blob: %w", err)
		}
		slog.Info("downloaded layer", "layer", fmt.Sprintf("%d/%d", i+1, len(manifest.Layers)), "digest", layer.Digest, "bytes", len(blob))
		if err := img.save(layerPath(layer.Digest), blob); err != nil {
			return Image{}, fmt.Errorf("cannot store layer blob: %w", err)
		}
	}

	slog.Info("pulled image", "name", img.Name, "tag", img.Tag, "config", manifest.Config.Digest, "layers", len(manifest.Layers))
	return *img, nil
}
