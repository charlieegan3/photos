package models

import (
	"time"
)

// Device represents data for devices used to take photos such as cameras and
// phones
type Device struct {
	ID      int
	Name    string
	IconURL string

	CreatedAt time.Time
	UpdatedAt time.Time
}
