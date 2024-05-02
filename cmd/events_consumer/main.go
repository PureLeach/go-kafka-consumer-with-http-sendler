package main

import (
	"events_consumer/internal/config"
	"events_consumer/internal/kafka"
	"fmt"
	"reflect"
)

func main() {
	// Загрузка конфигурации
	cfg := config.ConfigLoad()
	fmt.Printf("cfg %#v\n", cfg)
	fmt.Println("Тип переменной cfg:", reflect.TypeOf(cfg))

	// Пример использования переменной из конфигурации
	fmt.Printf("Address: %s\n", cfg.KAFKA_BOOTSTRAP_SERVER)
	kafka.ConsumerMain(cfg) // Вызов функции ConsumerMain() из пакета kafka
}
