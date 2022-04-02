package models

import "time"

type Location struct {
	ID        int
	Name      string
	Slug      string `json:"-"`
	Latitude  float64
	Longitude float64

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
