package types

import "time"

// Post represents a completed / downloaded post json file
type Post struct {
	Caption    string `json:"caption"`
	Code       string `json:"code"`
	Dimensions struct {
		Height int64 `json:"height"`
		Width  int64 `json:"width"`
	} `json:"dimensions"`
	DisplayURL string `json:"display_url"`
	ID         string `json:"id"`
	IsVideo    bool   `json:"is_video"`
	Location   struct {
		HasPublicPage bool   `json:"has_public_page"`
		ID            string `json:"id"`
		Name          string `json:"name"`
		Slug          string `json:"slug"`
	} `json:"location"`
	MediaURL  string   `json:"media_url"`
	PostURL   string   `json:"post_url"`
	Tags      []string `json:"tags"`
	Timestamp int64    `json:"timestamp"`
}

// FullID computes the full id from the date and instagram ID
func (c *Post) FullID() string {
	time := time.Unix(c.Timestamp, 0)
	return time.Format("2006-01-02") + "-" + c.ID
}
