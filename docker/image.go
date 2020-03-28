package docker

import (
	"context"
	"time"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"

	"github.com/JosephSalisbury/staler/stale"
)

// Image is a Staler for Docker images.
type Image struct {
	expiry time.Duration

	client *docker.Client
}

// NewImage returns a new Staler for Docker images.
func NewImage(expiry time.Duration) (*Image, error) {
	client, err := docker.NewEnvClient()
	if err != nil {
		return nil, err
	}

	image := &Image{
		expiry: expiry,

		client: client,
	}

	return image, nil
}

// Expiry returns the maximum age of Docker images.
func (i *Image) Expiry() time.Duration {
	return i.expiry
}

// List returns all Docker images.
func (i *Image) List() ([]stale.Item, error) {
	images, err := i.client.ImageList(
		context.Background(),
		types.ImageListOptions{
			All: true,
		},
	)
	if err != nil {
		return nil, err
	}

	items := []stale.Item{}
	for _, image := range images {
		item := stale.Item{
			ID:  image.ID,
			Age: time.Unix(image.Created, 0),
		}

		items = append(items, item)
	}

	return items, nil
}

// Delete deletes the specified Docker image.
func (i *Image) Delete(item stale.Item) error {
	_, err := i.client.ImageRemove(
		context.Background(),
		item.ID,
		types.ImageRemoveOptions{
			Force:         true,
			PruneChildren: true,
		},
	)

	return err
}

func (c *Image) String() string {
	return "Docker images"
}
