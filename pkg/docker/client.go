package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/vnFuhung2903/vcs-sms/entities"
)

type IDockerClient interface {
	Create(ctx context.Context, name string) (*entities.Container, error)
	IsRunning(ctx context.Context, id string) (bool, error)
}

type DockerClient struct {
	client *client.Client
}

func (c *DockerClient) Create(ctx context.Context, name string) (*entities.Container, error) {
	con, err := c.client.ContainerCreate(ctx, &container.Config{
		Image: "nginx",
	}, &container.HostConfig{}, &network.NetworkingConfig{}, nil, name)
	if err != nil {
		return nil, err
	}

	if err := c.client.ContainerStart(ctx, con.ID, container.StartOptions{}); err != nil {
		return nil, err
	}

	inspect, err := c.client.ContainerInspect(ctx, con.ID)
	if err != nil {
		return nil, err
	}

	network, ok := inspect.NetworkSettings.Networks["bridge"]
	if !ok {
		return nil, fmt.Errorf("cannot inspect container's adress")
	}

	var status string
	if inspect.State.Running {
		status = "ON"
	} else {
		status = "OFF"
	}
	return &entities.Container{
		ContainerId:   con.ID,
		ContainerName: name,
		Ipv4:          network.IPAddress,
		Status:        entities.ContainerStatus(status),
	}, nil
}

func (c *DockerClient) IsRunning(ctx context.Context, id string) (bool, error) {
	inspect, err := c.client.ContainerInspect(ctx, id)
	if err != nil {
		return false, err
	}

	isRunning := inspect.State.Running
	return isRunning, nil
}
