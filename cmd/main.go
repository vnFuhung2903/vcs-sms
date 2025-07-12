package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/vnFuhung2903/vcs-sms/api"
	"github.com/vnFuhung2903/vcs-sms/infrastructures/databases"
	"github.com/vnFuhung2903/vcs-sms/pkg/docker"
	"github.com/vnFuhung2903/vcs-sms/pkg/env"
	"github.com/vnFuhung2903/vcs-sms/pkg/logger"
	"github.com/vnFuhung2903/vcs-sms/usecases/repositories"
	"github.com/vnFuhung2903/vcs-sms/usecases/services"
)

func main() {
	env, err := env.LoadEnv(".")
	if err != nil {
		log.Fatalf("Failed to retrieve env: %v", err)
	}

	logger, err := logger.LoadLogger(env.LoggerEnv)
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}

	postgresDb := databases.ConnectPostgresDb(env.PostgresEnv)
	esClient := databases.ConnectESDb()
	redisClient := databases.ConnectRedis()
	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		log.Fatalf("Failed to create docker client: %v", err)
	}

	containerRepository := repositories.NewContainerRepository(postgresDb)
	userRepository := repositories.NewUserRepository(postgresDb)

	authService := services.NewAuthService(redisClient, logger)
	containerService := services.NewContainerService(containerRepository, logger)
	healthcheckService := services.NewHealthcheckService(containerRepository, dockerClient, esClient, logger)
	reportService := services.NewReportService(logger)
	userService := services.NewUserService(userRepository, logger)

	containerHandler := api.NewContainerHandler(containerService)
	reportHandler := api.NewReportHandler(containerService, healthcheckService, reportService)
	userHandler := api.NewUserHandler(authService, userService)

	r := gin.Default()
	containerHandler.SetupRoutes(r)
	reportHandler.SetupRoutes(r)
	userHandler.SetupRoutes(r)
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run container: %v", err)
	}
}
