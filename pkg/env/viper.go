package env

import "github.com/spf13/viper"

type AuthEnv struct {
	JWTSecret string `mapstructure:"JWT_SECRET_KEY"`
}

type GomailEnv struct {
	MailUsername string `mapstructure:"MAIL_USERNAME"`
	MailPassword string `mapstructure:"MAIL_PASSWORD"`
}

type PostgresEnv struct {
	PostgresUser     string `mapstructure:"POSTGRES_USER"`
	PostgresPassword string `mapstructure:"POSTGRES_PASSWORD"`
	PostgresName     string `mapstructure:"POSTGRES_NAME"`
}

type LoggerEnv struct {
	Level      string `mapstructure:"ZAP_LEVEL"`
	FilePath   string `mapstructure:"ZAP_FILEPATH"`
	MaxSize    int    `mapstructure:"ZAP_MAXSIZE"`
	MaxAge     int    `mapstructure:"ZAP_MAXAGE"`
	MaxBackups int    `mapstructure:"ZAP_MAXBACKUPS"`
}

type Env struct {
	AuthEnv     AuthEnv
	GomailEnv   GomailEnv
	PostgresEnv PostgresEnv
	LoggerEnv   LoggerEnv
}

func LoadEnv(path string) (*Env, error) {
	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var authEnv AuthEnv
	var gomailEnv GomailEnv
	var loggerEnv LoggerEnv
	var postgresEnv PostgresEnv

	if err := v.Unmarshal(&authEnv); err != nil {
		return nil, err
	}
	if err := v.Unmarshal(&gomailEnv); err != nil {
		return nil, err
	}
	if err := v.Unmarshal(&loggerEnv); err != nil {
		return nil, err
	}
	if err := v.Unmarshal(&postgresEnv); err != nil {
		return nil, err
	}
	return &Env{
		AuthEnv:     authEnv,
		GomailEnv:   gomailEnv,
		PostgresEnv: postgresEnv,
		LoggerEnv:   loggerEnv,
	}, nil
}
