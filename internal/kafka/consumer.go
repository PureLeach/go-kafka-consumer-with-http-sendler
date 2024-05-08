package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"

	"events_consumer/internal/config"
	"events_consumer/internal/models"
	"events_consumer/internal/utils"

	"github.com/IBM/sarama"
)

type Consumer struct {
}

func (*Consumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (*Consumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func ConsumerMain(cfg *config.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	setupConsumerGroup(ctx, cfg)
}

// setupConsumerGroup sets up a Sarama consumer group to consume Kafka messages
// and handle them with the Consumer struct.
func setupConsumerGroup(ctx context.Context, cfg *config.Config) {

	consumerGroup, err := initializeConsumerGroup(cfg)
	if err != nil {
		log.Printf("initialization error: %v", err)
		return
	}
	defer consumerGroup.Close()

	// Create a new Consumer struct
	consumer := &Consumer{}

	// Loop until the context is done or the consumer returns an error
	for {
		err = consumerGroup.Consume(ctx, []string{cfg.KAFKA_TOPIC}, consumer)
		if err != nil {
			log.Printf("error from consumer: %v", err)
		}
		if ctx.Err() != nil {
			return
		}
	}
}

func initializeConsumerGroup(cfg *config.Config) (sarama.ConsumerGroup, error) {
	fmt.Printf("Initializing kafka bootstrap server = %s, topic = %s, consumer group = %s\n", cfg.KAFKA_BOOTSTRAP_SERVER, cfg.KAFKA_TOPIC, cfg.KAFKA_CONSUMER_GROUP)
	config := sarama.NewConfig()

	consumerGroup, err := sarama.NewConsumerGroup(
		[]string{cfg.KAFKA_BOOTSTRAP_SERVER}, cfg.KAFKA_CONSUMER_GROUP, config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize consumer group: %w", err)
	}

	return consumerGroup, nil
}

func (consumer *Consumer) ConsumeClaim(
	sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	cfg := config.ConfigLoad()
	client := utils.CreateClient()

	fmt.Println("start listening topic for messages")

	for msg := range claim.Messages() {

		var kafkaMessage models.KafkaMessage
		err := json.Unmarshal(msg.Value, &kafkaMessage)
		if err != nil {
			log.Printf("Ошибка при парсинге сообщения: msg = %s, err = %v\n", msg.Value, err)
			continue
		}

		cacheCoreVehicleId := utils.CacheMain.Get(kafkaMessage.CloudVehicleID)
		if cacheCoreVehicleId != nil {
			coreVehicleId := cacheCoreVehicleId.Value()
			fmt.Printf("coreVehicleId: %#v Type: %v\n", coreVehicleId, reflect.TypeOf(coreVehicleId))
			sendVstRequest(coreVehicleId, kafkaMessage, client, cfg)
		} else {
			fmt.Println("Не нашли элемент в кэше")
		}

		sess.MarkMessage(msg, "")
	}
	return nil
}

func sendVstRequest(coreVehicleId string, kafkaMessage models.KafkaMessage, client *http.Client, cfg *config.Config) error {

	event := models.VehicleStateUpdateRequest{
		CoreVehicleId:                      coreVehicleId,
		Longitude:                          kafkaMessage.Longitude,
		Latitude:                           kafkaMessage.Latitude,
		Altitude:                           kafkaMessage.Altitude,
		Satellites:                         kafkaMessage.Satellites,
		HighResolutionTotalVehicleDistance: int(kafkaMessage.HighResolutionTotalVehicleDistance),
		Ts:                                 kafkaMessage.Ts,
		Speed:                              kafkaMessage.Speed,
		FuelLevel:                          kafkaMessage.FuelLevel,
		BatteryLevel:                       kafkaMessage.BatteryLevel,
	}

	// Преобразуем структуру в JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		log.Fatalf("Ошибка при маршализации JSON: %v", err)
	}
	fmt.Printf("jsonData: %#v Type: %v\n", string(jsonData), reflect.TypeOf(jsonData))

	// Создаем PATCH-запрос
	req, err := http.NewRequest("PATCH", cfg.VEHICLE_STATE_SERVICE_URL, strings.NewReader(string(jsonData)))
	if err != nil {
		log.Fatalf("Ошибка при создании PATCH-запроса: %v", err)
	}

	// Устанавливаем заголовок Content-Type
	req.Header.Set("Content-Type", "application/json")

	// Отправляем запрос
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	// Получите код статуса ответа
	statusCode := resp.StatusCode
	fmt.Println("Response Status Code:", statusCode)

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Ошибка при чтении ответа: %v", err)
	}

	// Выводим тело ответа
	log.Println("Ответ от сервера:", string(body))
	return nil
}
