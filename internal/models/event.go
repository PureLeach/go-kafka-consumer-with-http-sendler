package models

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Notification map[string]interface{}

// 	Msg1 string `string:"msg1"`
// 	Msg2 string `string:"msg2"`
// 	// From    User   `json:"from"`
// 	// To      User   `json:"to"`
// 	// Message string `json:"message"`
// }

type KafkaMessage struct {
	VehicleID                          string  `json:"vehicle_id"`
	TraceID                            string  `json:"trace_id"`
	Longitude                          float64 `json:"longitude"`
	Latitude                           float64 `json:"latitude"`
	Altitude                           int     `json:"altitude"`
	Satellites                         int     `json:"satellites"`
	HighResolutionTotalVehicleDistance int     `json:"high_resolution_total_vehicle_distance"`
	CurrentMileage                     int     `json:"currentMileage"`
	Ts                                 int64   `json:"ts"`
	Speed                              int     `json:"speed"`
	FuelLevel                          int     `json:"fuelLevel"`
	BatteryLevel                       float64 `json:"batteryLevel"`
}

type TelemetryEventRequest struct {
	Id                                 string
	Timestamp                          int64
	KmclBlockIccid                     string
	KmclVehicleId                      string
	KmclTraceId                        string
	EventHeaderHash                    string
	Longitude                          float64
	Latitude                           float64
	Altitude                           int
	Satellites                         int
	HighResolutionTotalVehicleDistance int
	CurrentMileage                     int
	Ts                                 int64
	Speed                              int
	FuelLevel                          int
	BatteryLevel                       float64
}
