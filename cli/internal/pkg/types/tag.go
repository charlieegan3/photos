package types

// Tag represents a collection of posts with a given tag
type Tag struct {
	Name string `json:"name"`

	Posts []Post `json:"posts"`
}
