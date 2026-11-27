package registry

import "fmt"

// ImageIndex represents the structure of an image index (manifest list) in the Docker registry.
type ImageIndex struct {
	SchemaVersion int                  `json:"schemaVersion"`
	MediaType     string               `json:"mediaType"`
	Manifests     []ManifestDescriptor `json:"manifests"`
}

// ManifestDescriptor represents an individual manifest entry in the image index, including its digest, media type, platform information, and size.
type ManifestDescriptor struct {
	Descriptor
	Annotations map[string]string `json:"annotations,omitempty"`
	Platform    ManifestPlatform  `json:"platform"`
}

// ManifestPlatform represents the platform information for a manifest entry, including architecture, operating system, and optional variant.
type ManifestPlatform struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
	Variant      string `json:"variant,omitempty"`
}

type ImageManifest struct {
	SchemaVersion int          `json:"schemaVersion"`
	MediaType     string       `json:"mediaType"`
	Config        Descriptor   `json:"config"`
	Layers        []Descriptor `json:"layers"`
}

type Descriptor struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
}

func (idx *ImageIndex) SelectPlatform(os, arch string) (*ManifestDescriptor, error) {
	for i := range idx.Manifests {
		manifest := &idx.Manifests[i]

		if manifest.Platform.OS == os &&
			manifest.Platform.Architecture == arch {
			return manifest, nil
		}
	}

	return nil, fmt.Errorf("cannot find platform %s/%s", os, arch)
}
