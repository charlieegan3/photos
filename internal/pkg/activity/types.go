package activity

import "time"

type Point struct {
	Timestamp time.Time `json:"timestamp"`

	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`

	Accuracy         float64 `json:"accuracy"`
	VerticalAccuracy float64 `json:"vertical_accuracy"`

	Velocity float64 `json:"velocity"`
}
