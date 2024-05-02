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

	"events_consumer/internal/config"
	"events_consumer/internal/models"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
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
	for msg := range claim.Messages() {

		var kafkaMessage models.KafkaMessage
		err := json.Unmarshal(msg.Value, &kafkaMessage)
		if err != nil {
			log.Printf("Ошибка при парсинге сообщения: msg = %s, err = %v", msg.Value, err)
			continue
		}

		event := models.VehicleStateUpdateRequest{
			CoreVehicleId:                      uuid.New().String(),
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

		// Преобразуем структуру в JSON
		jsonData, err := json.Marshal(event)
		if err != nil {
			log.Fatalf("Ошибка при маршализации JSON: %v", err)
		}
		fmt.Printf("jsonData: %#v Type: %v\n", string(jsonData), reflect.TypeOf(jsonData))

		// Создайте новый Transport с отключенной проверкой SSL
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		// Создайте клиент HTTP с настроенным Transport
		client := &http.Client{Transport: tr}

		// Создайте новый запрос GET с параметрами
		req, err := http.NewRequest("GET", "https://api.id.public.qa.core.dpkapp.ru/vehicles_id/vehicles/", nil)
		if err != nil {
			fmt.Println("Error creating GET request:", err)
		}

		// Добавьте заголовок "X-Api-Key"
		req.Header.Set("X-Api-Key", "a961554f9abc712de5d974fba22151fb")

		// Добавьте параметр "page"
		q := req.URL.Query()
		page := "1"
		q.Add("page", page)
		req.URL.RawQuery = q.Encode()

		// Выполните запрос
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending GET request:", err)
		}
		defer resp.Body.Close()

		// Прочитайте тело ответа, если это необходимо
		// Например, для прочтения JSON ответа можно использовать json.Decoder

		// Получите код статуса ответа
		statusCode := resp.StatusCode
		fmt.Println("Response Status Code:", statusCode)

		// Читаем тело ответа
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Ошибка при чтении ответа: %v", err)
		}

		// Выводим тело ответа
		// log.Println("Ответ от сервера:", string(body))

		var coreResponse models.CoreResponse
		coreErr := json.Unmarshal(body, &coreResponse)
		if coreErr != nil {
			log.Printf("Ошибка при парсинге сообщения: msg = %s, err = %v", string(body), coreErr)
		}
		// fmt.Printf("coreResponse: %#v Type: %v\n", coreResponse, reflect.TypeOf(coreResponse))

		// Проход по всем элементам Items
		// idCore := coreResponse.Result.Items[0].ID
		for _, item := range coreResponse.Result.Items {
			fmt.Println("ID:", item.ID)
			// Здесь можно добавить вывод других полей
			fmt.Println("---------")
		}

		// // Создаем PATCH-запрос
		// req2, err := http.NewRequest("PATCH", "https://api.dev.vlm.dpkapp.ru/vst/states/", strings.NewReader(string(jsonData)))
		// if err != nil {
		// 	log.Fatalf("Ошибка при создании PATCH-запроса: %v", err)
		// }

		// // Устанавливаем заголовок Content-Type
		// req.Header.Set("Content-Type", "application/json")

		// // Отправляем запрос
		// resp2, err := client.Do(req2)
		// if err != nil {
		// 	log.Fatalf("Ошибка при отправке запроса: %v", err)
		// }
		// defer resp2.Body.Close()

		// // Получите код статуса ответа
		// statusCode2 := resp2.StatusCode
		// fmt.Println("Response Status Code:", statusCode2)

		// // Читаем тело ответа
		// body2, err := io.ReadAll(resp.Body)
		// if err != nil {
		// 	log.Fatalf("Ошибка при чтении ответа: %v", err)
		// }

		// // Выводим тело ответа
		// log.Println("Ответ от сервера:", string(body2))

		sess.MarkMessage(msg, "")
	}
	return nil
}
