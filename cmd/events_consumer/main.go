package main

import (
	"events_consumer/internal/config"
	"fmt"
)

func main() {
	// Загрузка конфигурации
	cfg := config.ConfigLoad()

	// Пример использования переменной из конфигурации
	fmt.Printf("Address: %s\n", cfg.KAFKA_BOOTSTRAP_SERVER)
}
