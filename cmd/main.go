package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/vnFuhung2903/vcs-sms/api"
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/infrastructures/databases"
	"github.com/vnFuhung2903/vcs-sms/pkg/docker"
	"github.com/vnFuhung2903/vcs-sms/pkg/env"
	"github.com/vnFuhung2903/vcs-sms/pkg/logger"
	"github.com/vnFuhung2903/vcs-sms/pkg/middlewares"
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
	postgresDb.AutoMigrate(&entities.Container{}, &entities.User{})
	esClient := databases.ConnectESDb()
	redisClient := databases.ConnectRedis()
	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		log.Fatalf("Failed to create docker client: %v", err)
	}
	jwtMiddleware := middlewares.NewJWTMiddleware(redisClient, env.AuthEnv)

	containerRepository := repositories.NewContainerRepository(postgresDb)
	userRepository := repositories.NewUserRepository(postgresDb)

	containerService := services.NewContainerService(containerRepository, logger)
	healthcheckService := services.NewHealthcheckService(containerRepository, dockerClient, esClient, logger)
	reportService := services.NewReportService(logger, env.GomailEnv)
	userService := services.NewUserService(userRepository, logger)

	containerHandler := api.NewContainerHandler(containerService, jwtMiddleware)
	reportHandler := api.NewReportHandler(containerService, healthcheckService, reportService, jwtMiddleware)
	userHandler := api.NewUserHandler(userService, jwtMiddleware)

	r := gin.Default()
	containerHandler.SetupRoutes(r)
	reportHandler.SetupRoutes(r)
	userHandler.SetupRoutes(r)
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run container: %v", err)
	}
}
