package types

import (
	"errors"
)

// RawLocation contains the raw data that comes back from the gql endpoint
type RawLocation struct {
	Graphql struct {
		Location struct {
			AddressJSON string `json:"address_json"`
			Blurb       string `json:"blurb"`
			Directory   struct {
				City struct {
					ID   string `json:"id"`
					Name string `json:"name"`
					Slug string `json:"slug"`
				} `json:"city"`
				Country struct {
					ID   string `json:"id"`
					Name string `json:"name"`
					Slug string `json:"slug"`
				} `json:"country"`
			} `json:"directory"`
			EdgeLocationToMedia struct {
				Count int64 `json:"count"`
				Edges []struct {
					Node struct {
						AccessibilityCaption interface{} `json:"accessibility_caption"`
						CommentsDisabled     bool        `json:"comments_disabled"`
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
						ID      string `json:"id"`
						IsVideo bool   `json:"is_video"`
						Owner   struct {
							ID string `json:"id"`
						} `json:"owner"`
						Shortcode          string `json:"shortcode"`
						TakenAtTimestamp   int64  `json:"taken_at_timestamp"`
						ThumbnailResources []struct {
							ConfigHeight int64  `json:"config_height"`
							ConfigWidth  int64  `json:"config_width"`
							Src          string `json:"src"`
						} `json:"thumbnail_resources"`
						ThumbnailSrc   string `json:"thumbnail_src"`
						VideoViewCount int64  `json:"video_view_count"`
					} `json:"node"`
				} `json:"edges"`
				PageInfo struct {
					EndCursor   string `json:"end_cursor"`
					HasNextPage bool   `json:"has_next_page"`
				} `json:"page_info"`
			} `json:"edge_location_to_media"`
			EdgeLocationToTopPosts struct {
				Count int64 `json:"count"`
				Edges []struct {
					Node struct {
						AccessibilityCaption interface{} `json:"accessibility_caption"`
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
						ID      string `json:"id"`
						IsVideo bool   `json:"is_video"`
						Owner   struct {
							ID string `json:"id"`
						} `json:"owner"`
						Shortcode          string `json:"shortcode"`
						TakenAtTimestamp   int64  `json:"taken_at_timestamp"`
						ThumbnailResources []struct {
							ConfigHeight int64  `json:"config_height"`
							ConfigWidth  int64  `json:"config_width"`
							Src          string `json:"src"`
						} `json:"thumbnail_resources"`
						ThumbnailSrc   string `json:"thumbnail_src"`
						VideoViewCount int64  `json:"video_view_count"`
					} `json:"node"`
				} `json:"edges"`
				PageInfo struct {
					EndCursor   interface{} `json:"end_cursor"`
					HasNextPage bool        `json:"has_next_page"`
				} `json:"page_info"`
			} `json:"edge_location_to_top_posts"`
			HasPublicPage    bool    `json:"has_public_page"`
			ID               string  `json:"id"`
			Lat              float64 `json:"lat"`
			Lng              float64 `json:"lng"`
			Name             string  `json:"name"`
			Phone            string  `json:"phone"`
			PrimaryAliasOnFb string  `json:"primary_alias_on_fb"`
			ProfilePicURL    string  `json:"profile_pic_url"`
			Slug             string  `json:"slug"`
			Website          string  `json:"website"`
		} `json:"location"`
	} `json:"graphql"`
	LoggingPageID                    string `json:"logging_page_id"`
	PhotosAndVideosHeader            bool   `json:"photos_and_videos_header"`
	RecentPicturesAndVideosSubheader bool   `json:"recent_pictures_and_videos_subheader"`
	TopImagesAndVideosSubheader      bool   `json:"top_images_and_videos_subheader"`
}

// ToLocation returns the formatted location ready for saving
func (l *RawLocation) ToLocation() (Location, error) {
	location := l.Graphql.Location

	if location.ID == "" {
		return Location{}, errors.New("location in gql response was missing ID")
	}

	return Location{
		ID:   location.ID,
		Lat:  location.Lat,
		Long: location.Lng,
		Name: location.Name,
		Slug: location.Slug,
	}, nil
}
