package image

import (
	"fmt"

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

	_img, err := r.GetImage(img.Name, img.Tag)
	if err != nil {
		return Image{}, fmt.Errorf("error pulling image: %v", err)
	}

	fmt.Printf("Pulled image: %s:%s from %s\n", img.Name, img.Tag, _img)
	return *img, nil
}
