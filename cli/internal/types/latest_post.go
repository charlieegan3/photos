package types

// LatestPost describes a post on the users profile page
type LatestPost struct {
	Typename             string `json:"__typename"`
	AccessibilityCaption string `json:"accessibility_caption"`
	CommentsDisabled     bool   `json:"comments_disabled"`
	Dimensions           struct {
		Height int64 `json:"height"`
		Width  int64 `json:"width"`
	} `json:"dimensions"`
	DisplayURL  string `json:"display_url"`
	EdgeLikedBy struct {
		Count int64 `json:"count"`
	} `json:"edge_liked_by"`
	EdgeMediaPreviewLike struct {
		Count int64 `json:"count"`
	} `json:"edge_media_preview_like"`
	EdgeMediaToCaption struct {
		Edges []struct {
			Node struct {
				Text string `json:"text"`
			} `json:"node"`
		} `json:"edges"`
	} `json:"edge_media_to_caption"`
	EdgeMediaToComment struct {
		Count int64 `json:"count"`
	} `json:"edge_media_to_comment"`
	FactCheckInformation   interface{} `json:"fact_check_information"`
	FactCheckOverallRating interface{} `json:"fact_check_overall_rating"`
	GatingInfo             interface{} `json:"gating_info"`
	ID                     string      `json:"id"`
	IsVideo                bool        `json:"is_video"`
	Location               struct {
		HasPublicPage bool   `json:"has_public_page"`
		ID            string `json:"id"`
		Name          string `json:"name"`
		Slug          string `json:"slug"`
	} `json:"location"`
	MediaPreview string `json:"media_preview"`
	Owner        struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	} `json:"owner"`
	Shortcode          string `json:"shortcode"`
	TakenAtTimestamp   int64  `json:"taken_at_timestamp"`
	ThumbnailResources []struct {
		ConfigHeight int64  `json:"config_height"`
		ConfigWidth  int64  `json:"config_width"`
		Src          string `json:"src"`
	} `json:"thumbnail_resources"`
	ThumbnailSrc string `json:"thumbnail_src"`
}
