package models

import "time"

// Post represents a post to be shared on the site, linking a media item and
// various metadata
type Post struct {
	ID int

	Description string

	InstagramCode string

	PublishDate time.Time

	IsDraft bool

	CreatedAt time.Time
	UpdatedAt time.Time

	MediaID    int
	LocationID int
}
