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
					Node struct {
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
					} `json:"node"`
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
