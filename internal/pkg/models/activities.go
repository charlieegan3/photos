package models

import "time"

type Activity struct {
	ID int64

	Title string

	Description string

	StartTime time.Time
	EndTime   time.Time

	ImporterID int64
	CallerID   int64

	CreatedAt time.Time
	UpdatedAt time.Time
}
