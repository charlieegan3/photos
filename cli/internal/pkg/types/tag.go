package types

// TagIndexItem stores an item in a summary list of tags for the index page
type TagIndexItem struct {
	Name           string `json:"name"`
	MostRecentPost string `json:"most_recent"`
	Count          int    `json:"count"`
}

// Tag represents a collection of posts with a given tag
type Tag struct {
	Name string `json:"name"`

	Posts []Post `json:"posts"`
}
