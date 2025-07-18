package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/vnFuhung2903/vcs-sms/entities"
)

type IDockerClient interface {
	Create(ctx context.Context, name string, imageName string) (*container.CreateResponse, error)
	Start(ctx context.Context, containerID string) error
	GetStatus(ctx context.Context, containerID string) (entities.ContainerStatus, error)
	GetIpv4(ctx context.Context, containerID string) (string, error)
	Stop(ctx context.Context, containerID string) error
	Delete(ctx context.Context, containerID string) error
}

type DockerClient struct {
	client *client.Client
}

func NewDockerClient() (IDockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &DockerClient{
		client: cli,
	}, nil
}

func (c *DockerClient) Create(ctx context.Context, name string, imageName string) (*container.CreateResponse, error) {
	if err := c.PullImage(ctx, imageName); err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}

	con, err := c.client.ContainerCreate(ctx, &container.Config{
		Image: imageName,
	}, nil, nil, nil, name)
	return &con, err
}

func (c *DockerClient) Start(ctx context.Context, containerId string) error {
	return c.client.ContainerStart(ctx, containerId, container.StartOptions{})
}

func (c *DockerClient) GetStatus(ctx context.Context, containerId string) (entities.ContainerStatus, error) {
	inspect, err := c.client.ContainerInspect(ctx, containerId)
	if err != nil {
		return entities.ContainerOff, err
	}

	var status string
	if inspect.State.Running {
		status = "ON"
	} else {
		status = "OFF"
	}
	return entities.ContainerStatus(status), nil
}

func (c *DockerClient) GetIpv4(ctx context.Context, containerId string) (string, error) {
	inspect, err := c.client.ContainerInspect(ctx, containerId)
	if err != nil {
		return "", err
	}

	for _, network := range inspect.NetworkSettings.Networks {
		if network.IPAddress != "" {
			return network.IPAddress, nil
		}
	}
	return "", fmt.Errorf("no IP address found for container %s", containerId)
}

func (c *DockerClient) Stop(ctx context.Context, containerId string) error {
	return c.client.ContainerStop(ctx, containerId, container.StopOptions{})
}

func (c *DockerClient) Delete(ctx context.Context, containerId string) error {
	return c.client.ContainerRemove(ctx, containerId, container.RemoveOptions{
		Force: true,
	})
}

func (c *DockerClient) PullImage(ctx context.Context, refStr string) error {
	resp, err := c.client.ImagePull(ctx, refStr, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer resp.Close()
	return nil
}
