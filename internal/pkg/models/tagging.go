package models

import "time"

// Tagging links posts and tags
type Tagging struct {
	ID int

	CreatedAt time.Time
	UpdatedAt time.Time

	PostID int
	TagID  int
}
