package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/vnFuhung2903/vcs-sms/api"
	"github.com/vnFuhung2903/vcs-sms/infrastructures/databases"
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

	logger, err := logger.LoadLogger(env.LoggerEvn)
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}

	postgresDb := databases.ConnectPostgresDb(env.PostgresEnv)
	containerRepository := repositories.NewContainerRepository(postgresDb)
	containerService := services.NewContainerService(containerRepository, logger)
	containerHandler := api.NewContainerHandler(containerService)

	r := gin.Default()
	containerHandler.SetupRoutes(r)
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run container: %v", err)
	}
}
