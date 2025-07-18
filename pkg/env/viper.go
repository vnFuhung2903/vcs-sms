package env

import (
	"errors"

	"github.com/spf13/viper"
)

type AuthEnv struct {
	JWTSecret string `mapstructure:"JWT_SECRET_KEY"`
}

type ElasticsearchEnv struct {
	ElasticsearchAddress string `mapstructure:"ELASTICSEARCH_ADDRESS"`
}

type GomailEnv struct {
	MailUsername string `mapstructure:"MAIL_USERNAME"`
	MailPassword string `mapstructure:"MAIL_PASSWORD"`
}

type PostgresEnv struct {
	PostgresHost     string `mapstructure:"POSTGRES_HOST"`
	PostgresUser     string `mapstructure:"POSTGRES_USER"`
	PostgresPassword string `mapstructure:"POSTGRES_PASSWORD"`
	PostgresName     string `mapstructure:"POSTGRES_NAME"`
	PostgresPort     string `mapstructure:"POSTGRES_PORT"`
}

type RedisEnv struct {
	RedisAddress  string `mapstructure:"REDIS_ADDRESS"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDb       int    `mapstructure:"REDIS_DB"`
}

type LoggerEnv struct {
	Level      string `mapstructure:"ZAP_LEVEL"`
	FilePath   string `mapstructure:"ZAP_FILEPATH"`
	MaxSize    int    `mapstructure:"ZAP_MAXSIZE"`
	MaxAge     int    `mapstructure:"ZAP_MAXAGE"`
	MaxBackups int    `mapstructure:"ZAP_MAXBACKUPS"`
}

type Env struct {
	AuthEnv          AuthEnv
	GomailEnv        GomailEnv
	ElasticsearchEnv ElasticsearchEnv
	PostgresEnv      PostgresEnv
	RedisEnv         RedisEnv
	LoggerEnv        LoggerEnv
}

func LoadEnv(path string) (*Env, error) {
	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigName(".env")
	v.SetConfigType("env")

	v.SetDefault("ELASTICSEARCH_ADDRESS", "http://localhost:9200")
	v.SetDefault("POSTGRES_HOST", "localhost")
	v.SetDefault("POSTGRES_USER", "postgres")
	v.SetDefault("POSTGRES_PASSWORD", "postgres")
	v.SetDefault("POSTGRES_NAME", "postgres")
	v.SetDefault("POSTGRES_PORT", "5432")
	v.SetDefault("REDIS_ADDRESS", "localhost:6379")
	v.SetDefault("REDIS_PASSWORD", "")
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("ZAP_LEVEL", "info")
	v.SetDefault("ZAP_FILEPATH", "./logs/app.log")
	v.SetDefault("ZAP_MAXSIZE", 100)
	v.SetDefault("ZAP_MAXAGE", 10)
	v.SetDefault("ZAP_MAXBACKUPS", 30)

	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var authEnv AuthEnv
	var elasticsearchEnv ElasticsearchEnv
	var gomailEnv GomailEnv
	var loggerEnv LoggerEnv
	var postgresEnv PostgresEnv
	var redisEnv RedisEnv

	if err := v.Unmarshal(&authEnv); err != nil || authEnv.JWTSecret == "" {
		err = errors.New("auth environment variables are empty")
		return nil, err
	}
	if err := v.Unmarshal(&elasticsearchEnv); err != nil || elasticsearchEnv.ElasticsearchAddress == "" {
		err = errors.New("elasticsearch environment variables are empty")
		return nil, err
	}
	if err := v.Unmarshal(&gomailEnv); err != nil || gomailEnv.MailUsername == "" {
		err = errors.New("gomail environment variables are empty")
		return nil, err
	}
	if err := v.Unmarshal(&loggerEnv); err != nil {
		return nil, err
	}
	if err := v.Unmarshal(&postgresEnv); err != nil || postgresEnv.PostgresUser == "" || postgresEnv.PostgresName == "" || postgresEnv.PostgresHost == "" || postgresEnv.PostgresPort == "" {
		err = errors.New("posgres environment variables are empty")
		return nil, err
	}
	if err := v.Unmarshal(&redisEnv); err != nil || redisEnv.RedisAddress == "" {
		err = errors.New("redis environment variables are empty")
		return nil, err
	}
	return &Env{
		AuthEnv:          authEnv,
		ElasticsearchEnv: elasticsearchEnv,
		GomailEnv:        gomailEnv,
		PostgresEnv:      postgresEnv,
		RedisEnv:         redisEnv,
		LoggerEnv:        loggerEnv,
	}, nil
}
