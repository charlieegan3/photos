package types

// Profile represents the full struct that comes back from the Graphql endpoint
// curl https://www.instagram.com/username/?__a=1 | gojson
type Profile struct {
	Graphql struct {
		User struct {
			Biography              string      `json:"biography"`
			BlockedByViewer        bool        `json:"blocked_by_viewer"`
			BusinessCategoryName   interface{} `json:"business_category_name"`
			CategoryID             interface{} `json:"category_id"`
			ConnectedFbPage        interface{} `json:"connected_fb_page"`
			CountryBlock           bool        `json:"country_block"`
			EdgeFelixVideoTimeline struct {
				Count    int64         `json:"count"`
				Edges    []interface{} `json:"edges"`
				PageInfo struct {
					EndCursor   interface{} `json:"end_cursor"`
					HasNextPage bool        `json:"has_next_page"`
				} `json:"page_info"`
			} `json:"edge_felix_video_timeline"`
			EdgeFollow struct {
				Count int64 `json:"count"`
			} `json:"edge_follow"`
			EdgeFollowedBy struct {
				Count int64 `json:"count"`
			} `json:"edge_followed_by"`
			EdgeMediaCollections struct {
				Count    int64         `json:"count"`
				Edges    []interface{} `json:"edges"`
				PageInfo struct {
					EndCursor   interface{} `json:"end_cursor"`
					HasNextPage bool        `json:"has_next_page"`
				} `json:"page_info"`
			} `json:"edge_media_collections"`
			EdgeMutualFollowedBy struct {
				Count int64         `json:"count"`
				Edges []interface{} `json:"edges"`
			} `json:"edge_mutual_followed_by"`
			EdgeOwnerToTimelineMedia struct {
				Count int64 `json:"count"`
				Edges []struct {
					Node LatestPost `json:"node"`
				} `json:"edges"`
				PageInfo struct {
					EndCursor   string `json:"end_cursor"`
					HasNextPage bool   `json:"has_next_page"`
				} `json:"page_info"`
			} `json:"edge_owner_to_timeline_media"`
			EdgeSavedMedia struct {
				Count    int64         `json:"count"`
				Edges    []interface{} `json:"edges"`
				PageInfo struct {
					EndCursor   interface{} `json:"end_cursor"`
					HasNextPage bool        `json:"has_next_page"`
				} `json:"page_info"`
			} `json:"edge_saved_media"`
			ExternalURL            string      `json:"external_url"`
			ExternalURLLinkshimmed string      `json:"external_url_linkshimmed"`
			FollowedByViewer       bool        `json:"followed_by_viewer"`
			FollowsViewer          bool        `json:"follows_viewer"`
			FullName               string      `json:"full_name"`
			HasArEffects           bool        `json:"has_ar_effects"`
			HasBlockedViewer       bool        `json:"has_blocked_viewer"`
			HasChannel             bool        `json:"has_channel"`
			HasRequestedViewer     bool        `json:"has_requested_viewer"`
			HighlightReelCount     int64       `json:"highlight_reel_count"`
			ID                     string      `json:"id"`
			IsBusinessAccount      bool        `json:"is_business_account"`
			IsJoinedRecently       bool        `json:"is_joined_recently"`
			IsPrivate              bool        `json:"is_private"`
			IsVerified             bool        `json:"is_verified"`
			OverallCategoryName    interface{} `json:"overall_category_name"`
			ProfilePicURL          string      `json:"profile_pic_url"`
			ProfilePicURLHd        string      `json:"profile_pic_url_hd"`
			RequestedByViewer      bool        `json:"requested_by_viewer"`
			RestrictedByViewer     interface{} `json:"restricted_by_viewer"`
			Username               string      `json:"username"`
		} `json:"user"`
	} `json:"graphql"`
	LoggingPageID         string      `json:"logging_page_id"`
	ShowFollowDialog      bool        `json:"show_follow_dialog"`
	ShowSuggestedProfiles bool        `json:"show_suggested_profiles"`
	ToastContentOnLoad    interface{} `json:"toast_content_on_load"`
}
