package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
	"github.com/vnFuhung2903/vcs-sms/api"
	_ "github.com/vnFuhung2903/vcs-sms/docs"
	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/infrastructures/databases"
	"github.com/vnFuhung2903/vcs-sms/interfaces"
	"github.com/vnFuhung2903/vcs-sms/pkg/docker"
	"github.com/vnFuhung2903/vcs-sms/pkg/env"
	"github.com/vnFuhung2903/vcs-sms/pkg/logger"
	"github.com/vnFuhung2903/vcs-sms/pkg/middlewares"
	"github.com/vnFuhung2903/vcs-sms/usecases/repositories"
	"github.com/vnFuhung2903/vcs-sms/usecases/services"
	"github.com/vnFuhung2903/vcs-sms/workers"
)

// @title VCS SMS API
// @version 1.0
// @description Container Management System API
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	env, err := env.LoadEnv(".")
	if err != nil {
		log.Fatalf("Failed to retrieve env: %v", err)
	}

	logger, err := logger.LoadLogger(env.LoggerEnv)
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}

	postgresDb, err := databases.ConnectPostgresDb(env.PostgresEnv)
	if err != nil {
		log.Fatalf("Failed to create docker client: %v", err)
	}
	postgresDb.AutoMigrate(&entities.Container{}, &entities.User{})

	esRawClient, err := databases.NewElasticsearchFactory(env.ElasticsearchEnv).ConnectElasticsearch()
	if err != nil {
		log.Fatalf("Failed to create docker client: %v", err)
	}
	esClient := interfaces.NewElasticsearchClient(esRawClient)

	redisRawClient := databases.NewRedisFactory(env.RedisEnv).ConnectRedis()
	redisClient := interfaces.NewRedisClient(redisRawClient)

	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		log.Fatalf("Failed to create docker client: %v", err)
	}
	jwtMiddleware := middlewares.NewJWTMiddleware(redisClient, env.AuthEnv)

	containerRepository := repositories.NewContainerRepository(postgresDb)
	userRepository := repositories.NewUserRepository(postgresDb)

	containerService := services.NewContainerService(containerRepository, dockerClient, logger)
	healthcheckService := services.NewHealthcheckService(containerRepository, dockerClient, esClient, logger)
	reportService := services.NewReportService(logger, env.GomailEnv)
	userService := services.NewUserService(userRepository, logger)

	containerHandler := api.NewContainerHandler(containerService, jwtMiddleware)
	reportHandler := api.NewReportHandler(containerService, healthcheckService, reportService, jwtMiddleware)
	userHandler := api.NewUserHandler(userService, jwtMiddleware)

	esWorker := workers.NewHealthcheckWorker(
		dockerClient,
		containerService,
		healthcheckService,
		logger,
		10*time.Second,
	)
	esWorker.Start(1)

	r := gin.Default()
	containerHandler.SetupRoutes(r)
	reportHandler.SetupRoutes(r)
	userHandler.SetupRoutes(r)
	r.GET("/swagger/*any", swagger.WrapHandler(swaggerFiles.Handler))

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		logger.Info("Shutting down...")
		esWorker.Stop()
		os.Exit(0)
	}()
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run container: %v", err)
	}
}
