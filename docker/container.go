package docker

import (
	"context"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"

	"github.com/JosephSalisbury/staler/stale"
)

// Container is a Staler for exited Docker containers.
type Container struct {
	expiry time.Duration

	client *docker.Client
}

// NewContainer returns a new Staler for exited Docker containers.
func NewContainer(expiry time.Duration) (*Container, error) {
	client, err := docker.NewEnvClient()
	if err != nil {
		return nil, err
	}

	container := &Container{
		expiry: expiry,

		client: client,
	}

	return container, nil
}

// Expiry returns the maximum age of exited Docker containers.
func (c *Container) Expiry() time.Duration {
	return c.expiry
}

// List returns all exited Docker containers.
func (c *Container) List() ([]stale.Item, error) {
	containers, err := c.client.ContainerList(
		context.Background(),
		types.ContainerListOptions{
			All:     true,
			Filters: filters.NewArgs(filters.Arg("status", "exited")),
		},
	)
	if err != nil {
		return nil, err
	}

	items := []stale.Item{}
	for _, container := range containers {
		item := stale.Item{
			ID:  container.ID,
			Age: time.Unix(container.Created, 0),
		}

		items = append(items, item)
	}

	return items, nil
}

// Delete deletes the specified Docker container.
func (c *Container) Delete(item stale.Item) error {
	err := c.client.ContainerRemove(
		context.Background(),
		item.ID,
		types.ContainerRemoveOptions{},
	)

	return err
}

func (c *Container) String() string {
	return "Docker containers"
}
