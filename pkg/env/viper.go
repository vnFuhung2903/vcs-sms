package env

import "github.com/spf13/viper"

type PostgresEnv struct {
	PostgresUser     string `mapstructure:"POSTGRES_USER"`
	PostgresPassword string `mapstructure:"POSTGRES_PASSWORD"`
	PostgresName     string `mapstructure:"POSTGRES_NAME"`
}

type LoggerEvn struct {
	Level      string `mapstructure:"ZAP_LEVEL"`
	FilePath   string `mapstructure:"ZAP_FILEPATH"`
	MaxSize    int    `mapstructure:"ZAP_MAXSIZE"`
	MaxAge     int    `mapstructure:"ZAP_MAXAGE"`
	MaxBackups int    `mapstructure:"ZAP_MAXBACKUPS"`
}

type Env struct {
	PostgresEnv PostgresEnv
	LoggerEvn   LoggerEvn
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

	var loggerEnv LoggerEvn
	var postgresEnv PostgresEnv
	if err := v.Unmarshal(&loggerEnv); err != nil {
		return nil, err
	}
	if err := v.Unmarshal(&postgresEnv); err != nil {
		return nil, err
	}
	return &Env{
		PostgresEnv: postgresEnv,
		LoggerEvn:   loggerEnv,
	}, nil
}
