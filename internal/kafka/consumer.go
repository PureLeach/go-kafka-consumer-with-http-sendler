package kafka

import (
	"context"
	"crypto/tls"
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
	// Создайте новый Transport с отключенной проверкой SSL
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Создайте клиент HTTP с настроенным Transport
	client := &http.Client{Transport: tr}

	for msg := range claim.Messages() {

		var kafkaMessage models.KafkaMessage
		err := json.Unmarshal(msg.Value, &kafkaMessage)
		if err != nil {
			log.Printf("Ошибка при парсинге сообщения: msg = %s, err = %v", msg.Value, err)
			continue
		}

		x := utils.CacheMain.Get(kafkaMessage.CloudVehicleID)
		fmt.Printf("x: %#v Type: %v\n", x, reflect.TypeOf(x))
		if x != nil {
			coreVehicleId := x.Value()
			fmt.Printf("coreVehicleId: %#v Type: %v\n", coreVehicleId, reflect.TypeOf(coreVehicleId))
			sendVstRequest(coreVehicleId, kafkaMessage, client)
		} else {
			fmt.Println("Не нашли элемент в кэше")
		}

		// h := sha1.New()
		// h.Write(msg.Value)
		// sha1Hash := h.Sum(nil)

		// eventClick := models.TelemetryEventRequest{
		// 	Id:                                 uuid.New().String(),
		// 	Timestamp:                          msg.Timestamp.UnixMicro(),
		// 	KmclBlockIccid:                     string(msg.Key),
		// 	EventHeaderHash:                    hex.EncodeToString(sha1Hash),
		// 	KmclVehicleId:                      kafkaMessage.VehicleID,
		// 	KmclTraceId:                        kafkaMessage.TraceID,
		// 	Longitude:                          kafkaMessage.Longitude,
		// 	Latitude:                           kafkaMessage.Latitude,
		// 	Altitude:                           kafkaMessage.Altitude,
		// 	Satellites:                         kafkaMessage.Satellites,
		// 	HighResolutionTotalVehicleDistance: kafkaMessage.HighResolutionTotalVehicleDistance,
		// 	CurrentMileage:                     kafkaMessage.CurrentMileage,
		// 	Ts:                                 kafkaMessage.Ts,
		// 	Speed:                              kafkaMessage.Speed,
		// 	FuelLevel:                          kafkaMessage.FuelLevel,
		// 	BatteryLevel:                       kafkaMessage.BatteryLevel,
		// }
		// fmt.Printf("eventClick: %#v Type: %v\n", eventClick, reflect.TypeOf(eventClick))

		sess.MarkMessage(msg, "")
	}
	return nil
}

func sendVstRequest(coreVehicleId string, kafkaMessage models.KafkaMessage, client *http.Client) error {

	event := models.VehicleStateUpdateRequest{
		// CoreVehicleId:                      uuid.New().String(),
		CoreVehicleId:                      coreVehicleId,
		Longitude:                          kafkaMessage.Longitude,
		Latitude:                           kafkaMessage.Latitude,
		Altitude:                           kafkaMessage.Altitude,
		Satellites:                         kafkaMessage.Satellites,
		HighResolutionTotalVehicleDistance: kafkaMessage.HighResolutionTotalVehicleDistance,
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
	req2, err := http.NewRequest("PATCH", "https://api.dev.vlm.dpkapp.ru/vst/states/", strings.NewReader(string(jsonData)))
	if err != nil {
		log.Fatalf("Ошибка при создании PATCH-запроса: %v", err)
	}

	// Устанавливаем заголовок Content-Type
	req2.Header.Set("Content-Type", "application/json")

	// Отправляем запрос
	resp2, err := client.Do(req2)
	if err != nil {
		log.Fatalf("Ошибка при отправке запроса: %v", err)
	}
	defer resp2.Body.Close()

	// Получите код статуса ответа
	statusCode2 := resp2.StatusCode
	fmt.Println("Response Status Code:", statusCode2)

	// Читаем тело ответа
	body2, err := io.ReadAll(resp2.Body)
	if err != nil {
		log.Fatalf("Ошибка при чтении ответа: %v", err)
	}

	// Выводим тело ответа
	log.Println("Ответ от сервера:", string(body2))
	return nil
}
