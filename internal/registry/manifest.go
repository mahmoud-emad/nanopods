package registry

// ImageIndex represents the structure of an image index (manifest list) in the Docker registry.
type ImageIndex struct {
	SchemaVersion int                  `json:"schemaVersion"`
	MediaType     string               `json:"mediaType"`
	Manifests     []ManifestDescriptor `json:"manifests"`
}

// ManifestDescriptor represents an individual manifest entry in the image index, including its digest, media type, platform information, and size.
type ManifestDescriptor struct {
	Annotations map[string]string `json:"annotations,omitempty"`
	Digest      string            `json:"digest"`
	MediaType   string            `json:"mediaType"`
	Platform    ManifestPlatform  `json:"platform"`
	Size        int64             `json:"size"`
}

// ManifestPlatform represents the platform information for a manifest entry, including architecture, operating system, and optional variant.
type ManifestPlatform struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
	Variant      string `json:"variant,omitempty"`
}
