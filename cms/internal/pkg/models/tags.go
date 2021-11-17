package models

import "time"

// Tag is a single tag associated with a post
type Tag struct {
	ID     int
	Name   string
	Hidden bool

	CreatedAt time.Time
	UpdatedAt time.Time
}
