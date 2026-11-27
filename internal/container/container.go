package container

import (
	"fmt"
	"log/slog"

	"github.com/mahmoud-emad/nanopods/internal/image"
	"github.com/mahmoud-emad/nanopods/internal/utils"
)

// Container represents a lightweight isolated workload with its own process namespace and resource limits.
type Container struct {
	ID    string
	Name  string
	PID   int
	Image *image.Image
}

// New creates a new container instance with the given name, image name, and tag. It generates a random ID for the container.
func New(name, img, tag string) *Container {
	return &Container{
		Name:  name,
		ID:    utils.RandomString(10),
		Image: image.New(img, tag),
	}
}

// Run simulates running the container by printing its details. In a real implementation, this would involve setting up namespaces, cgroups, and executing the container's process.
func (cn *Container) Run() error {
	slog.Info("starting container", "name", cn.Name, "id", cn.ID)
	img, imgErr := cn.Image.Pull()
	if imgErr != nil {
		return fmt.Errorf("error running container: %v", imgErr)
	}

	slog.Info("running container", "name", cn.Name, "image", img.Name, "tag", img.Tag)
	return nil
}
