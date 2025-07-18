package databases

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/pkg/env"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabasesSuite struct {
	suite.Suite
	db                *gorm.DB
	postgresContainer testcontainers.Container
	ctx               context.Context
}

func (suite *DatabasesSuite) SetupSuite() {
	suite.ctx = context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForSQL("5432/tcp", "postgres", func(host string, port nat.Port) string {
			return fmt.Sprintf("host=%s port=%s user=testuser password=testpass dbname=testdb sslmode=disable", host, port.Port())
		}).WithStartupTimeout(60 * time.Second),
	}

	postgresContainer, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	suite.Require().NoError(err)
	suite.postgresContainer = postgresContainer
}

func (suite *DatabasesSuite) TearDownSuite() {
	if suite.postgresContainer != nil {
		suite.postgresContainer.Terminate(suite.ctx)
	}
}

func (suite *DatabasesSuite) SetupTest() {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	suite.NoError(err)
	suite.db = gormDB
}

func (suite *DatabasesSuite) TearDownTest() {
	sqlDB, err := suite.db.DB()
	assert.NoError(suite.T(), err)
	sqlDB.Close()
}

func TestDatabasesSuite(t *testing.T) {
	suite.Run(t, new(DatabasesSuite))
}

func (suite *DatabasesSuite) TestConnectPostgresDb() {
	host, err := suite.postgresContainer.Host(suite.ctx)
	suite.NoError(err)

	mappedPort, err := suite.postgresContainer.MappedPort(suite.ctx, "5432")
	suite.NoError(err)

	pgEnv := env.PostgresEnv{
		PostgresHost:     host,
		PostgresUser:     "testuser",
		PostgresPassword: "testpass",
		PostgresName:     "testdb",
		PostgresPort:     mappedPort.Port(),
	}

	db, err := ConnectPostgresDb(pgEnv)
	suite.NoError(err)
	suite.NotNil(db)

	var result int
	err = db.Raw("SELECT 1").Scan(&result).Error
	suite.NoError(err)
	suite.Equal(1, result)

	var tableExists bool
	err = db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'containers')").Scan(&tableExists).Error
	suite.NoError(err)
	suite.True(tableExists)

	container := entities.Container{
		ContainerId:   "test-container-id",
		Status:        entities.ContainerOn,
		ContainerName: "test-container",
		Ipv4:          "192.168.1.100",
	}

	err = db.Create(&container).Error
	suite.NoError(err)

	var retrievedContainer entities.Container
	err = db.Where("container_id = ?", "test-container-id").First(&retrievedContainer).Error
	suite.NoError(err)
	suite.Equal(container.ContainerId, retrievedContainer.ContainerId)
	suite.Equal(container.Status, retrievedContainer.Status)
	suite.Equal(container.ContainerName, retrievedContainer.ContainerName)
	suite.Equal(container.Ipv4, retrievedContainer.Ipv4)

	sqlDB, err := db.DB()
	suite.NoError(err)
	sqlDB.Close()
}

func (suite *DatabasesSuite) TestConnectPostgresDbInvalidDsn() {
	invalidEnv := env.PostgresEnv{
		PostgresHost:     "localhost",
		PostgresUser:     "invalid_user",
		PostgresPassword: "invalid_pass",
		PostgresName:     "invalid_db",
		PostgresPort:     "5432",
	}

	db, err := ConnectPostgresDb(invalidEnv)
	suite.Error(err)
	suite.Nil(db)
	suite.Contains(err.Error(), "connect")
}

func (suite *DatabasesSuite) TestConnectPostgresDbInvalidPort() {
	invalidEnv := env.PostgresEnv{
		PostgresHost:     "postgres",
		PostgresUser:     "testuser",
		PostgresPassword: "testpass",
		PostgresName:     "testdb",
		PostgresPort:     "99999",
	}

	db, err := ConnectPostgresDb(invalidEnv)
	suite.Error(err)
	suite.Nil(db)
	suite.Contains(err.Error(), "parse")
}

func (suite *DatabasesSuite) TestConnectRedis() {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	suite.NoError(err)
	defer func() { _ = redisContainer.Terminate(ctx) }()

	env := env.RedisEnv{
		RedisAddress:  "localhost:6379",
		RedisPassword: "",
		RedisDb:       0,
	}

	redisFactory := NewRedisFactory(env)
	redisClient := redisFactory.ConnectRedis()
	suite.NotNil(redisClient)
}

func (suite *DatabasesSuite) TestConnectElasticsearch() {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "docker.elastic.co/elasticsearch/elasticsearch:8.17.4",
		ExposedPorts: []string{"9200/tcp"},
		WaitingFor:   wait.ForListeningPort("9200/tcp"),
	}
	elasticsearchContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	suite.NoError(err)
	defer func() { _ = elasticsearchContainer.Terminate(ctx) }()

	env := env.ElasticsearchEnv{
		ElasticsearchAddress: "http://localhost:9200",
	}

	elasticsearchFactory := NewElasticsearchFactory(env)
	elasticsearchClient, err := elasticsearchFactory.ConnectElasticsearch()
	suite.NotNil(elasticsearchClient)
	suite.NoError(err)
}
