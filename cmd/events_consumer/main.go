package main

import (
	"events_consumer/internal/config"
	"events_consumer/internal/integrations"
	"events_consumer/internal/kafka"
)

func main() {
	cfg := config.ConfigLoad()

	integrations.LoadCoreVehicle(cfg)

	// consumersData := map[string]string{
	// 	"key1": "value1",
	// 	"key2": "value2",
	// 	// Добавьте другие элементы карты по вашему усмотрению
	// }
	// consumer = kafka.Consumer()
	// consumer.store.Add(cache)

	// myStruct := kafka.NewMyStruct()

	// myStruct.SetValue("key1", 10)
	// myStruct.SetValue("key2", 20)

	// &kafka.Consumer{
	// 	Data: "Some data", // Добавляем данные в структуру
	// }

	kafka.ConsumerMain(cfg)
	// x := utils.CacheMain.Get("7771")
	// fmt.Printf("x: %#v Type: %v\n", x, reflect.TypeOf(x))
	// if x != nil {
	// 	y := x.Value()
	// 	fmt.Printf("y: %#v Type: %v\n", y, reflect.TypeOf(y))
	// }

}
