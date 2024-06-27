package clickhouse

import (
	"context"
	"events_consumer/internal/config"
	"events_consumer/internal/models"

	clickhouse "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"
)

// setupClickHouse подключается к ClickHouse и создаёт таблицу, если она не существует
func SetupClickHouse(cfg *config.Config) (clickhouse.Conn, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"10.42.0.27:9000"},
		Auth: clickhouse.Auth{
			Database: "telemetry",
			Username: "default",
			Password: "5KC3BSvS",
		},
		// Указываем настройки соединения
		Debug: false,
	})
	if err != nil {
		return nil, err
	}

	// Создаём таблицу (если она ещё не существует)
	createTable := `
		CREATE TABLE IF NOT EXISTS telemetry.raw_events_extended (
			id UUID,
			vehicle_id String,
			trace_id Nullable(String),
			longitude Nullable(Float64),
			latitude Nullable(Float64),
			ts Nullable(UInt64),
			time String,
			speed Nullable(UInt64),
			vin String,
			INDEX idx_vehicle_time (vehicle_id, time) TYPE set(64) GRANULARITY 4,
			INDEX idx_vin_time (vin, time) TYPE set(64) GRANULARITY 4
		) ENGINE = MergeTree()
		ORDER BY (vehicle_id, vin, time, id)
		SETTINGS index_granularity = 8192
	`
	if err := conn.Exec(context.Background(), createTable); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

// InsertBatch записывает данные в ClickHouse
func InsertBatch(conn clickhouse.Conn, data []models.KafkaMessage) error {
	batch, err := conn.PrepareBatch(context.Background(), "INSERT INTO telemetry.raw_events_extended (id, vehicle_id, trace_id, longitude, latitude, ts, time, vin)")
	if err != nil {
		return err
	}

	for _, message := range data {
		if err := batch.Append(uuid.New().String(), message.CloudVehicleID, message.TraceID, message.Longitude, message.Latitude, message.Ts, message.Time, message.Vin); err != nil {
			return err
		}
	}

	return batch.Send()
}
