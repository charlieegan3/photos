package models

import (
	"time"
)

// Lens is a go struct representing a lens record in the rest of the application
type Lens struct {
	ID   int64
	Name string

	CreatedAt time.Time
	UpdatedAt time.Time
}
