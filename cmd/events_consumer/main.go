package main

import (
	"events_consumer/internal/config"
	"events_consumer/internal/kafka"
)

func main() {
	cfg := config.ConfigLoad()

	// integrations.LoadCoreVehicle(cfg)

	kafka.ConsumerMain(cfg)
}
