package models

import "time"

type Location struct {
	ID        int
	Name      string
	Slug      string
	Latitude  float32
	Longitude float32

	CreatedAt time.Time
	UpdatedAt time.Time
}
