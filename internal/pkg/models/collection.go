package models

import "time"

type Collection struct {
	ID int

	Title       string
	Description string

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
