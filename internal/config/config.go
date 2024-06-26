package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	KAFKA_BOOTSTRAP_SERVER string `env:"KAFKA_BOOTSTRAP_SERVER" env-default:"localhost:9093"`
	KAFKA_TOPIC            string `env:"KAFKA_TOPIC" env-required:"true"`
	KAFKA_CONSUMER_GROUP   string `env:"KAFKA_CONSUMER_GROUP" env-required:"true"`

	SECURITY_PROTOCOL    string `env:"SECURITY_PROTOCOL" env-required:"true"`
	SSL_CA_PATH          string `env:"SSL_CA_PATH" env-required:"false"`
	SSL_CERTIFICATE_PATH string `env:"SSL_CERTIFICATE_PATH" env-required:"false"`
	SSL_KEY_PATH         string `env:"SSL_KEY_PATH" env-required:"false"`

	CORE_HOST                 string `env:"CORE_HOST" env-required:"true"`
	CORE_VEHICLE_SERVICE_PATH string `env:"CORE_VEHICLE_SERVICE_PATH" env-required:"true"`
	CORE_API_KEY              string `env:"CORE_API_KEY" env-required:"true"`

	VEHICLE_STATE_SERVICE_URL string `env:"VEHICLE_STATE_SERVICE_URL" env-required:"true"`

	CLICKHOUSE_HOST     string `env:"CLICKHOUSE_HOST" env-required:"true"`
	CLICKHOUSE_PORT     int    `env:"CLICKHOUSE_PORT" env-required:"true"`
	CLICKHOUSE_DBNAME   string `env:"CLICKHOUSE_DBNAME" env-required:"true"`
	CLICKHOUSE_USERNAME string `env:"CLICKHOUSE_USERNAME" env-required:"true"`
	CLICKHOUSE_PASSWORD string `env:"CLICKHOUSE_PASSWORD" env-required:"true"`
}

func ConfigLoad() *Config {

	var cfg Config

	if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
		log.Fatalf("Configuration read failed: %s", err)
	}
	return &cfg
}
