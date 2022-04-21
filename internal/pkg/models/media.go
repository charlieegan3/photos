package models

import "time"

// Media represents a media item uploaded to the system
type Media struct {
	ID int

	Kind string

	Make  string
	Model string
	Lens  string

	TakenAt time.Time

	FNumber                 float64
	ExposureTimeNumerator   uint32
	ExposureTimeDenominator uint32
	ISOSpeed                int

	Latitude  float64
	Longitude float64
	Altitude  float64

	CreatedAt time.Time
	UpdatedAt time.Time

	DeviceID int

	LensID int64

	InstagramCode string
}
