package databases

import (
	"fmt"
	"log"

	"github.com/vnFuhung2903/vcs-sms/entities"
	"github.com/vnFuhung2903/vcs-sms/pkg/env"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectPostgresDb(env env.PostgresEnv) *gorm.DB {
	dsn := fmt.Sprintf("host=localhost user=%s password=%s dbname=%s port=5432 sslmode=disable", env.PostgresUser, env.PostgresPassword, env.PostgresName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Postgres connection error: %v", err)
	}

	err = db.AutoMigrate(&entities.Server{})
	if err != nil {
		log.Fatalf("Migrate error: %v", err)
	}
	return db
}
