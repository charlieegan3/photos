package models

import "time"

type Location struct {
	ID        int     `json:"Id"`
	Name      string  `json:"Name"`
	Slug      string  `json:"-"`
	Latitude  float64 `json:"Latitude"`
	Longitude float64 `json:"Longitude"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
