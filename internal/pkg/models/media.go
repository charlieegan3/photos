package models

import "time"

// Media represents a media item uploaded to the system
type Media struct {
	ID int

	Kind string

	Make  string
	Model string

	Lens        string
	FocalLength string

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

	DeviceID int64

	LensID int64

	InstagramCode string

	// UTCCorrect is true when the UTC time is known to be correct from the time of upload.
	// Older imported images will not have this set since the data was imported from Instagram
	UTCCorrect bool

	Width  int
	Height int

	DisplayOffset int
}
