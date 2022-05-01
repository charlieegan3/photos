package models

import "time"

type Point struct {
	ID int64

	Latitude  float64
	Longitude float64
	Altitude  float64

	Accuracy         float64
	VerticalAccuracy float64

	Velocity float64

	WasOffline bool

	ImporterID int64
	CallerID   int64
	ReasonID   int64

	ActivityID *int64

	CreatedAt time.Time
	UpdatedAt time.Time
}
