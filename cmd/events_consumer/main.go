package main

import (
	"events_consumer/internal/config"
	"events_consumer/internal/kafka"
	"fmt"
)

func main() {
	// Загрузка конфигурации
	cfg := config.ConfigLoad()

	// Пример использования переменной из конфигурации
	fmt.Printf("Address: %s\n", cfg.KAFKA_BOOTSTRAP_SERVER)
	kafka.ConsumerMain() // Вызов функции ConsumerMain() из пакета kafka
}
