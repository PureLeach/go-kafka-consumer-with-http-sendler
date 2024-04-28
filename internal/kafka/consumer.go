package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"events_consumer/internal/models"

	"github.com/IBM/sarama"
)

const (
	ConsumerGroup      = "notifications-group"
	ConsumerTopic      = "notifications"
	ConsumerPort       = ":8081"
	KafkaServerAddress = "localhost:9093"
)

// ====== NOTIFICATION STORAGE ======
// type UserNotifications map[string][]models.Notification

// type NotificationStore struct {
// 	data UserNotifications
// 	mu   sync.RWMutex
// }

// func (ns *NotificationStore) Add(userID string,
// 	notification models.Notification) {
// 	ns.mu.Lock()
// 	defer ns.mu.Unlock()
// 	ns.data[userID] = append(ns.data[userID], notification)
// }

// func (ns *NotificationStore) Get(userID string) []models.Notification {
// 	ns.mu.RLock()
// 	defer ns.mu.RUnlock()
// 	return ns.data[userID]
// }

// ============== KAFKA RELATED FUNCTIONS ==============
type Consumer struct {
	// store *NotificationStore
}

func (*Consumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (*Consumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (consumer *Consumer) ConsumeClaim(
	sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		// value := string(msg.Value)
		// fmt.Printf("value %#v\n", value)
		// userID := string(msg.Key)
		// fmt.Printf("userID %#v\n", userID)
		var notification models.Notification
		err := json.Unmarshal(msg.Value, &notification)
		if err != nil {
			log.Printf("failed to unmarshal notification: %v", err)
			continue
		}
		fmt.Printf("notification %#v\n", notification)
		// consumer.store.Add(userID, notification)
		sess.MarkMessage(msg, "")
	}
	return nil
}

func initializeConsumerGroup() (sarama.ConsumerGroup, error) {
	config := sarama.NewConfig()

	consumerGroup, err := sarama.NewConsumerGroup(
		[]string{KafkaServerAddress}, ConsumerGroup, config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize consumer group: %w", err)
	}

	return consumerGroup, nil
}

func setupConsumerGroup(ctx context.Context) {
	// func setupConsumerGroup(ctx context.Context, store *NotificationStore) {
	consumerGroup, err := initializeConsumerGroup()
	if err != nil {
		log.Printf("initialization error: %v", err)
	}
	defer consumerGroup.Close()

	consumer := &Consumer{
		// store: store,
	}

	for {
		err = consumerGroup.Consume(ctx, []string{ConsumerTopic}, consumer)
		if err != nil {
			log.Printf("error from consumer: %v", err)
		}
		if ctx.Err() != nil {
			return
		}
	}
}

func ConsumerMain() {
	// store := &NotificationStore{
	// 	data: make(UserNotifications),
	// }

	ctx, cancel := context.WithCancel(context.Background())
	setupConsumerGroup(ctx)
	// go setupConsumerGroup(ctx, store)
	defer cancel()

	fmt.Printf("Kafka CONSUMER (Group: %s) 👥📥 "+
		"started at http://localhost%s\n", ConsumerGroup, ConsumerPort)

}
