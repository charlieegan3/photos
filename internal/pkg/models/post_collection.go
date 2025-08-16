package models

// PostCollection links posts and collections.
type PostCollection struct {
	ID int

	PostID       int
	CollectionID int
}
