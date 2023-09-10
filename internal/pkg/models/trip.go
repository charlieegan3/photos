package models

import "time"

type Trip struct {
	ID int

	Title       string
	Description string

	StartDate time.Time
	EndDate   time.Time

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
