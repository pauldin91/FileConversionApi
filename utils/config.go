package utils

type Config struct {
	Environment       string `mapstructure:"ENVIRONMENT"`
	ConnString        string `mapstructure:"DBURL"`
	HttpServerAddress string `mapstructure:"HTTP_SERVER_ADDRESS"`
	CertificatePath   string `mapstructure:"CERTIFICATE_PATH"`
}
