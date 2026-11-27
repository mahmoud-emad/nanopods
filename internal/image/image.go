package image

import (
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

func (img *Image) Pull() (Image, error) {
	// we suould later change this to use a registry interface to support multiple registries, for now we will use the default registry.
	r := registry.New()

	slog.Info("pulling image", "name", img.Name, "tag", img.Tag, "id", img.ID)
	imgIdx, err := r.GetImage(img.Name, img.Tag)
	if err != nil {
		return Image{}, fmt.Errorf("error pulling image: %v", err)
	}

	// Find the manifest matching the host platform.
	// For now, let's assume linux/amd64. TODO: We need to support different platforms.
	platform, err := imgIdx.SelectPlatform("linux", "amd")
	if err != nil {
		return Image{}, fmt.Errorf("cannot select platform: %v", err)
	}

	slog.Info("selected platform", "digest", platform.Digest, "os", platform.Platform.OS, "arch", platform.Platform.Architecture)

	manifest, err := r.GetManifest(img.Name, platform.Digest)
	if err != nil {
		return Image{}, fmt.Errorf("cannot fetch image manifest: %v", err)
	}

	slog.Info("pulled image", "name", img.Name, "tag", img.Tag, "config", manifest.Config.Digest, "layers", len(manifest.Layers))
	return *img, nil
}
