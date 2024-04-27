package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	KAFKA_BOOTSTRAP_SERVER string `env:"KAFKA_BOOTSTRAP_SERVER" env-default:"localhost:9093"`
	KAFKA_TOPIC            string `env:"KAFKA_TOPIC" env-required:"true"`
	KAFKA_CONSUMER_GROUP   string `env:"KAFKA_CONSUMER_GROUP" env-required:"true"`
}

func ConfigLoad() *Config {

	var cfg Config

	if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
		log.Fatalf("Configuration read failed: %s", err)
	}
	return &cfg
}
