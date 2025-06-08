package env

import "github.com/spf13/viper"

type DatabaseEnv struct {
	PostgresUser       string `mapstructure:"POSTGRES_USER"`
	PostgresPassword   string `mapstructure:"POSTGRES_PASSWORD"`
	PostgresName       string `mapstructure:"POSTGRES_NAME"`
	KafkaBrokerAddress string `mapstructure:"KAFKA_BROKER_ADDRESS"`
}

type LoggerEvn struct {
	Level      string `mapstructure:"ZAP_LEVEL"`
	FilePath   string `mapstructure:"ZAP_FILEPATH"`
	MaxSize    int    `mapstructure:"ZAP_MAXSIZE"`
	MaxAge     int    `mapstructure:"ZAP_MAXAGE"`
	MaxBackups int    `mapstructure:"ZAP_MAXBACKUPS"`
}

type Env struct {
	DatabaseEnv DatabaseEnv
	LoggerEvn   LoggerEvn
}

func LoadEnv(path string) (env Env, err error) {
	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigName(".env")
	v.SetConfigType("env")

	v.AutomaticEnv()

	err = v.ReadInConfig()
	if err != nil {
		return
	}

	err = v.Unmarshal(&env)
	return
}
