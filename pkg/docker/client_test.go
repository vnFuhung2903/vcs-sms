package docker

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type DockerClientTestSuite struct {
	suite.Suite
	client IDockerClient
	ctx    context.Context
}

func (suite *DockerClientTestSuite) SetupTest() {
	suite.ctx = context.Background()

	client, err := NewDockerClient()
	suite.client = client
	suite.NoError(err)
}

func TestDockerClientTestSuite(t *testing.T) {
	suite.Run(t, new(DockerClientTestSuite))
}

func (suite *DockerClientTestSuite) TestContainerOnLifeCycle() {
	con, err := suite.client.Create(suite.ctx, "test-container", "nginx:1.28")
	suite.NoError(err)

	err = suite.client.Start(suite.ctx, con.ID)
	suite.NoError(err)

	_, err = suite.client.GetStatus(suite.ctx, con.ID)
	suite.NoError(err)
	_, err = suite.client.GetIpv4(suite.ctx, con.ID)
	suite.NoError(err)

	err = suite.client.Stop(suite.ctx, con.ID)
	suite.NoError(err)

	err = suite.client.Delete(suite.ctx, con.ID)
	suite.NoError(err)
}

func (suite *DockerClientTestSuite) TestContainerOffLifeCycle() {
	con, err := suite.client.Create(suite.ctx, "test-container", "nginx:1.28")
	suite.NoError(err)

	_, err = suite.client.GetStatus(suite.ctx, con.ID)
	suite.NoError(err)
	_, err = suite.client.GetIpv4(suite.ctx, con.ID)
	suite.Error(err)

	err = suite.client.Delete(suite.ctx, con.ID)
	suite.NoError(err)
}

func (suite *DockerClientTestSuite) TestPullImageInvalidImage() {
	dockerClient := suite.client.(*DockerClient)
	err := dockerClient.PullImage(suite.ctx, "invalid/non-existent-image:invalid-tag")
	suite.Error(err)
	suite.Contains(strings.ToLower(err.Error()), "failed to pull image")
}

func (suite *DockerClientTestSuite) TestCreateContainerInvalidImage() {
	_, err := suite.client.Create(suite.ctx, "test-container", "invalid/non-existent-image")
	suite.Error(err)
}

func (suite *DockerClientTestSuite) TestGetStatusNonExistentContainer() {
	_, err := suite.client.GetStatus(suite.ctx, "non-existent-container-id")
	suite.Error(err)
}

func (suite *DockerClientTestSuite) TestGetIpv4NonExistentContainer() {
	_, err := suite.client.GetIpv4(suite.ctx, "non-existent-container-id")
	suite.Error(err)
}

func (suite *DockerClientTestSuite) TestStartNonExistentContainer() {
	err := suite.client.Start(suite.ctx, "non-existent-container-id")
	suite.Error(err)
}

func (suite *DockerClientTestSuite) TestStopNonExistentContainer() {
	err := suite.client.Stop(suite.ctx, "non-existent-container-id")
	suite.Error(err)
}

func (suite *DockerClientTestSuite) TestDeleteNonExistentContainer() {
	err := suite.client.Delete(suite.ctx, "non-existent-container-id")
	suite.T().Logf("Delete non-existent container result: %v", err)
}
