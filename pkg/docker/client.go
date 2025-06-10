package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type IDockerClient interface {
	Create(ctx context.Context, name string) error
}

type DockerClient struct {
	client *client.Client
}

func (c *DockerClient) Create(ctx context.Context, name string) error {
	res, err := c.client.ContainerCreate(ctx, &container.Config{
		Image: "nginx",
	}, &container.HostConfig{}, &network.NetworkingConfig{}, nil, name)
	if err != nil {
		return err
	}

	if err := c.client.ContainerStart(ctx, res.ID, container.StartOptions{}); err != nil {
		return err
	}

	inspect, err := c.client.ContainerInspect(ctx, res.ID)
	if err != nil {
		return err
	}

	if network, ok := inspect.NetworkSettings.Networks["bridge"]; ok {
		fmt.Println("Container IPv4:", network.IPAddress)
	} else {
		return fmt.Errorf("cannot inspect container's adress")
	}
	return nil
}
