package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"events_consumer/internal/clickhouse"
	"events_consumer/internal/config"
	"events_consumer/internal/models"

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
		log.Fatalf("initialization error: %v", err)
		return
	}
	defer consumerGroup.Close()

	// Create a new Consumer struct
	consumer := &Consumer{}

	// Loop until the context is done or the consumer returns an error
	for {
		err = consumerGroup.Consume(ctx, []string{cfg.KAFKA_TOPIC}, consumer)
		if err != nil {
			log.Fatalf("Error from consumer: %v", err)
		}
		if ctx.Err() != nil {
			return
		}
	}
}

func NewTLSConfig(clientCertFile, clientKeyFile, caCertFile string) (*tls.Config, error) {
	tlsConfig := tls.Config{}

	// Load client cert
	cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	if err != nil {
		return &tlsConfig, err
	}
	tlsConfig.Certificates = []tls.Certificate{cert}

	// Load CA cert
	caCert, err := os.ReadFile(caCertFile)
	if err != nil {
		return &tlsConfig, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig.RootCAs = caCertPool

	// tlsConfig.BuildNameToCertificate()
	return &tlsConfig, err
}

func initializeConsumerGroup(cfg *config.Config) (sarama.ConsumerGroup, error) {
	log.Printf("Initializing kafka bootstrap server = %s, topic = %s, consumer group = %s\n", cfg.KAFKA_BOOTSTRAP_SERVER, cfg.KAFKA_TOPIC, cfg.KAFKA_CONSUMER_GROUP)

	// This can be used on test server if domain does not match cert:
	// tlsConfig.InsecureSkipVerify = true

	config := sarama.NewConfig()

	if cfg.SECURITY_PROTOCOL == "SSL" {
		tlsConfig, err := NewTLSConfig(cfg.SSL_CERTIFICATE_PATH,
			cfg.SSL_KEY_PATH,
			cfg.SSL_CA_PATH)
		if err != nil {
			log.Fatal(err)
		}
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	} else if cfg.SECURITY_PROTOCOL == "PLAINTEXT" {
		config.Net.TLS.Enable = false

	}

	consumerGroup, err := sarama.NewConsumerGroup(
		[]string{cfg.KAFKA_BOOTSTRAP_SERVER}, cfg.KAFKA_CONSUMER_GROUP, config)
	if err != nil {
		log.Fatal(err)
	}

	return consumerGroup, nil
}

func (consumer *Consumer) ConsumeClaim(
	sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	log.Println("Инициализация клиента ClickHouse")
	cfg := config.ConfigLoad()

	// Подключаемся к ClickHouse и создаём таблицу
	conn, err := clickhouse.SetupClickHouse(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	// client := utils.CreateClient()

	log.Println("start listening topic for messages")

	// Creating a channel with a fixed size to limit the number of simultaneous requests
	// requests := make(chan struct{}, 100) // for example, let's limit up to 100 requests at a time

	// Настраиваем буфер и канал для накопления данных
	batchSize := 10000
	// data := make([][]interface{}, 0, batchSize)
	data := make([]models.KafkaMessage, 0, batchSize)
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for msg := range claim.Messages() {
		var kafkaMessage models.KafkaMessage
		err := json.Unmarshal(msg.Value, &kafkaMessage)
		if err != nil {
			log.Printf("Error parsing the message: msg = %s, err = %v\n", msg.Value, err)
			continue
		}
		data = append(data, kafkaMessage)
		// Если накопили достаточный объём данных, записываем их в ClickHouse
		if len(data) >= batchSize {
			if err := clickhouse.InsertBatch(conn, data); err != nil {
				log.Println("Ошибка при записи данных:", err)
			}
			data = data[:0] // Очищаем буфер
			log.Println("Данные записаны в ClickHouse")
		}

		// x := kafkaMessage.CloudVehicleID
		// log.Printf("kafkaMessage.CloudVehicleID = %s\n", x)

		// cacheCoreVehicleId := utils.CacheMain.Get(kafkaMessage.CloudVehicleID)
		// if cacheCoreVehicleId != nil {
		// 	coreVehicleId := cacheCoreVehicleId.Value()

		// 	// Sending a request to the channel
		// 	requests <- struct{}{}
		// 	go func() {
		// 		sendVstRequest(coreVehicleId, kafkaMessage, client, cfg)
		// 		// After executing the request, we extract the signal from the channel
		// 		<-requests
		// 	}()
		// } else {
		// 	log.Println("The item was not found in the cache")
		// }

		sess.MarkMessage(msg, "")
	}
	return nil
}

func sendVstRequest(coreVehicleId string, kafkaMessage models.KafkaMessage, client *http.Client, cfg *config.Config) {
	event := models.VehicleStateUpdateRequest{
		CoreVehicleId: coreVehicleId,
		Longitude:     kafkaMessage.Longitude,
		Latitude:      kafkaMessage.Latitude,
		// Altitude:                           kafkaMessage.Altitude,
		// Satellites:                         kafkaMessage.Satellites,
		// HighResolutionTotalVehicleDistance: int(kafkaMessage.HighResolutionTotalVehicleDistance),
		Ts: kafkaMessage.Ts,
		// Speed: kafkaMessage.Speed,
		// FuelLevel:                          kafkaMessage.FuelLevel,
		// BatteryLevel:                       kafkaMessage.BatteryLevel,
	}

	jsonData, err := json.Marshal(event)
	if err != nil {
		log.Fatalf("Error during JSON marshalization: %v", err)
	}

	req, err := http.NewRequest("PATCH", cfg.VEHICLE_STATE_SERVICE_URL, strings.NewReader(string(jsonData)))
	if err != nil {
		log.Fatalf("Error when creating a PATCH request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending the request: %v", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	log.Println("Response Status Code:", statusCode)

}
