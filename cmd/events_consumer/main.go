package main

import (
	"events_consumer/internal/config"
	"events_consumer/internal/kafka"
)

func main() {
	cfg := config.ConfigLoad()
	kafka.ConsumerMain(cfg)
}
