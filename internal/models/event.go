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
	CloudVehicleID                     string  `json:"vehicle_id"`
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

type VehicleStateUpdateRequest struct {
	CoreVehicleId                      string  `json:"vehicle_id"`
	Longitude                          float64 `json:"longitude"`
	Latitude                           float64 `json:"latitude"`
	Altitude                           int     `json:"altitude"`
	Satellites                         int     `json:"satellites"`
	HighResolutionTotalVehicleDistance int     `json:"vehicle_distance"`
	Ts                                 int64   `json:"external_ts"`
	Speed                              int     `json:"speed"`
	FuelLevel                          int     `json:"fuel_level_percent"`
	BatteryLevel                       float64 `json:"car_battery_level"`
}

type TelemetryEventRequest struct {
	Id                                 string  `json:"id"`
	Timestamp                          int64   `json:"timestamp"`
	KmclBlockIccid                     string  `json:"kmcl_block_iccid"`
	KmclVehicleId                      string  `json:"kmcl_vehicle_id"`
	KmclTraceId                        string  `json:"kmcl_trace_id"`
	EventHeaderHash                    string  `json:"event_header_hash"`
	Longitude                          float64 `json:"longitude"`
	Latitude                           float64 `json:"latitude"`
	Altitude                           int     `json:"altitude"`
	Satellites                         int     `json:"satellites"`
	HighResolutionTotalVehicleDistance int     `json:"high_resolution_total_vehicle_distance"`
	CurrentMileage                     int     `json:"current_mileage"`
	Ts                                 int64   `json:"ts"`
	Speed                              int     `json:"speed"`
	FuelLevel                          int     `json:"fuel_level"`
	BatteryLevel                       float64 `json:"battery_level"`
}

type CoreResponse struct {
	Error      string `json:"error"`
	HTTPStatus int64  `json:"http_status"`
	Result     struct {
		Items []struct {
			BlockID        string        `json:"block_id"`
			Brand          string        `json:"brand"`
			CreatedDt      string        `json:"created_dt"`
			Creator        interface{}   `json:"creator"`
			ID             string        `json:"id"`
			KmclVehicleID  string        `json:"kmcl_vehicle_id"`
			LicenseNumber  string        `json:"license_number"`
			Model          string        `json:"model"`
			OrganizationID string        `json:"organization_id"`
			Status         interface{}   `json:"status"`
			UpdatedDt      string        `json:"updated_dt"`
			VehicleGroups  []interface{} `json:"vehicle_groups"`
			Vin            string        `json:"vin"`
			VinEngine      string        `json:"vin_engine"`
		} `json:"items"`
		Page  int64 `json:"page"`
		Pages int64 `json:"pages"`
		Size  int64 `json:"size"`
		Total int64 `json:"total"`
	} `json:"result"`
	Success bool `json:"success"`
}
