package kafka

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"events_consumer/internal/config"
	"events_consumer/internal/models"

	"github.com/google/uuid"

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
	for msg := range claim.Messages() {

		var kafkaMessage models.KafkaMessage
		err := json.Unmarshal(msg.Value, &kafkaMessage)
		if err != nil {
			log.Printf("Ошибка при парсинге сообщения: msg = %s, err = %v", msg.Value, err)
			continue
		}

		h := sha1.New()
		h.Write(msg.Value)
		sha1Hash := h.Sum(nil)

		event := models.TelemetryEventRequest{
			Id:                                 uuid.New().String(),
			Timestamp:                          msg.Timestamp.UnixMicro(),
			KmclBlockIccid:                     string(msg.Key),
			EventHeaderHash:                    hex.EncodeToString(sha1Hash),
			KmclVehicleId:                      kafkaMessage.VehicleID,
			KmclTraceId:                        kafkaMessage.TraceID,
			Longitude:                          kafkaMessage.Longitude,
			Latitude:                           kafkaMessage.Latitude,
			Altitude:                           kafkaMessage.Altitude,
			Satellites:                         kafkaMessage.Satellites,
			HighResolutionTotalVehicleDistance: kafkaMessage.HighResolutionTotalVehicleDistance,
			CurrentMileage:                     kafkaMessage.CurrentMileage,
			Ts:                                 kafkaMessage.Ts,
			Speed:                              kafkaMessage.Speed,
			FuelLevel:                          kafkaMessage.FuelLevel,
			BatteryLevel:                       kafkaMessage.BatteryLevel,
		}
		fmt.Printf("event: %#v Type: %v\n", event, reflect.TypeOf(event))

		sess.MarkMessage(msg, "")
	}
	return nil
}
