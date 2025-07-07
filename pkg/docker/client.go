package docker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/vnFuhung2903/vcs-sms/entities"
)

type IDockerClient interface {
	Create(ctx context.Context, name string, imageName string, refStr string) (*entities.Container, error)
	HealthCheck(ctx context.Context, id string) error
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

func (c *DockerClient) Create(ctx context.Context, name string, imageName string, refStr string) (*entities.Container, error) {
	if err := c.PullImage(ctx, refStr); err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}

	con, err := c.client.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Healthcheck: &container.HealthConfig{
			Test:     []string{"CMD-SHELL", "curl -f http://localhost/ || exit 1"},
			Interval: 10 * time.Second,
			Timeout:  2 * time.Second,
			Retries:  3,
		},
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

	networkSettings, ok := inspect.NetworkSettings.Networks["bridge"]
	if !ok {
		return nil, errors.New("cannot inspect container's address")
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
		Ipv4:          networkSettings.IPAddress,
		Status:        entities.ContainerStatus(status),
	}, nil
}

func (c *DockerClient) HealthCheck(ctx context.Context, id string) error {
	inspect, err := c.client.ContainerInspect(ctx, id)
	if err != nil {
		return err
	}
	networkInfo, ok := inspect.NetworkSettings.Networks["bridge"]
	if !ok {
		return errors.New("cannot find bridge network")
	}
	containerIP := networkInfo.IPAddress
	url := fmt.Sprintf("http://%s/", containerIP)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("API call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("healthcheck API returned status %d", resp.StatusCode)
	}
	return nil
}

func (c *DockerClient) Delete(ctx context.Context, containerId string) error {
	return c.client.ContainerRemove(ctx, containerId, container.RemoveOptions{
		Force: true,
	})
}

func (c *DockerClient) PullImage(ctx context.Context, refStr string) error {
	resp, err := c.client.ImagePull(ctx, refStr, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer resp.Close()
	return nil
}
