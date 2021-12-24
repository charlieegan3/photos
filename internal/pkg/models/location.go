package models

import "time"

type Location struct {
	ID        int
	Name      string
	Slug      string
	Latitude  float64
	Longitude float64

	CreatedAt time.Time
	UpdatedAt time.Time
}
