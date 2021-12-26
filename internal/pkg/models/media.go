package models

import "time"

// Media represents a media item uploaded to the system
type Media struct {
	ID int

	Kind string

	Make  string
	Model string

	TakenAt time.Time

	FNumber                 float64
	ShutterSpeed            float64
	ExposureTimeNumerator   uint32
	ExposureTimeDenominator uint32
	ISOSpeed                int

	Latitude  float64
	Longitude float64
	Altitude  float64

	CreatedAt time.Time
	UpdatedAt time.Time

	DeviceID int

	InstagramCode string
}
