package utils

import "github.com/spf13/viper"

type Config struct {
	Environment       string `mapstructure:"ENVIRONMENT"`
	ConnectionString  string `mapstructure:"CONNECTION_STRING"`
	MigrationLocation string `mapstructure:"MIGRATION_LOCATION"`
	HttpServerAddress string `mapstructure:"HTTP_SERVER_ADDRESS"`
	CertPath          string `mapstructure:"CERTIFICATE_PATH"`
	CertFile          string `mapstructure:"CERTIFICATE_FILE"`
	CertKey           string `mapstructure:"CERTIFICATE_KEY"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
