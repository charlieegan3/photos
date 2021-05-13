package types

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// RawPost represents the full struct that comes back from the Graphql endpoint
// curl https://www.instagram.com/p/B-xfipLA2Yz/?__a=1 | gojson
type RawPost struct {
	Graphql struct {
		ShortcodeMedia struct {
			Typename                    string `json:"__typename"`
			AccessibilityCaption        string `json:"accessibility_caption"`
			CaptionIsEdited             bool   `json:"caption_is_edited"`
			CommentingDisabledForViewer bool   `json:"commenting_disabled_for_viewer"`
			CommentsDisabled            bool   `json:"comments_disabled"`
			Dimensions                  struct {
				Height int64 `json:"height"`
				Width  int64 `json:"width"`
			} `json:"dimensions"`
			DisplayResources []struct {
				ConfigHeight int64  `json:"config_height"`
				ConfigWidth  int64  `json:"config_width"`
				Src          string `json:"src"`
			} `json:"display_resources"`
			DisplayURL              string `json:"display_url"`
			EdgeMediaPreviewComment struct {
				Count int64 `json:"count"`
				Edges []struct {
					Node struct {
						CreatedAt       int64 `json:"created_at"`
						DidReportAsSpam bool  `json:"did_report_as_spam"`
						EdgeLikedBy     struct {
							Count int64 `json:"count"`
						} `json:"edge_liked_by"`
						ID                  string `json:"id"`
						IsRestrictedPending bool   `json:"is_restricted_pending"`
						Owner               struct {
							ID            string `json:"id"`
							IsVerified    bool   `json:"is_verified"`
							ProfilePicURL string `json:"profile_pic_url"`
							Username      string `json:"username"`
						} `json:"owner"`
						Text           string `json:"text"`
						ViewerHasLiked bool   `json:"viewer_has_liked"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"edge_media_preview_comment"`
			EdgeMediaPreviewLike struct {
				Count int64         `json:"count"`
				Edges []interface{} `json:"edges"`
			} `json:"edge_media_preview_like"`
			EdgeMediaToCaption struct {
				Edges []struct {
					Node struct {
						Text string `json:"text"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"edge_media_to_caption"`
			EdgeMediaToHoistedComment struct {
				Edges []interface{} `json:"edges"`
			} `json:"edge_media_to_hoisted_comment"`
			EdgeMediaToParentComment struct {
				Count int64 `json:"count"`
				Edges []struct {
					Node struct {
						CreatedAt       int64 `json:"created_at"`
						DidReportAsSpam bool  `json:"did_report_as_spam"`
						EdgeLikedBy     struct {
							Count int64 `json:"count"`
						} `json:"edge_liked_by"`
						EdgeThreadedComments struct {
							Count    int64         `json:"count"`
							Edges    []interface{} `json:"edges"`
							PageInfo struct {
								EndCursor   interface{} `json:"end_cursor"`
								HasNextPage bool        `json:"has_next_page"`
							} `json:"page_info"`
						} `json:"edge_threaded_comments"`
						ID                  string `json:"id"`
						IsRestrictedPending bool   `json:"is_restricted_pending"`
						Owner               struct {
							ID            string `json:"id"`
							IsVerified    bool   `json:"is_verified"`
							ProfilePicURL string `json:"profile_pic_url"`
							Username      string `json:"username"`
						} `json:"owner"`
						Text           string `json:"text"`
						ViewerHasLiked bool   `json:"viewer_has_liked"`
					} `json:"node"`
				} `json:"edges"`
				PageInfo struct {
					EndCursor   interface{} `json:"end_cursor"`
					HasNextPage bool        `json:"has_next_page"`
				} `json:"page_info"`
			} `json:"edge_media_to_parent_comment"`
			EdgeMediaToSponsorUser struct {
				Edges []interface{} `json:"edges"`
			} `json:"edge_media_to_sponsor_user"`
			EdgeMediaToTaggedUser struct {
				Edges []interface{} `json:"edges"`
			} `json:"edge_media_to_tagged_user"`
			EdgeRelatedProfiles struct {
				Edges []interface{} `json:"edges"`
			} `json:"edge_related_profiles"`
			EdgeWebMediaToRelatedMedia struct {
				Edges []interface{} `json:"edges"`
			} `json:"edge_web_media_to_related_media"`
			FactCheckInformation   interface{} `json:"fact_check_information"`
			FactCheckOverallRating interface{} `json:"fact_check_overall_rating"`
			GatingInfo             interface{} `json:"gating_info"`
			HasRankedComments      bool        `json:"has_ranked_comments"`
			ID                     string      `json:"id"`
			IsAd                   bool        `json:"is_ad"`
			IsVideo                bool        `json:"is_video"`
			Location               struct {
				AddressJSON   string `json:"address_json"`
				HasPublicPage bool   `json:"has_public_page"`
				ID            string `json:"id"`
				Name          string `json:"name"`
				Slug          string `json:"slug"`
			} `json:"location"`
			MediaOverlayInfo interface{} `json:"media_overlay_info"`
			MediaPreview     string      `json:"media_preview"`
			Owner            struct {
				BlockedByViewer bool `json:"blocked_by_viewer"`
				EdgeFollowedBy  struct {
					Count int64 `json:"count"`
				} `json:"edge_followed_by"`
				EdgeOwnerToTimelineMedia struct {
					Count int64 `json:"count"`
				} `json:"edge_owner_to_timeline_media"`
				FollowedByViewer   bool        `json:"followed_by_viewer"`
				FullName           string      `json:"full_name"`
				HasBlockedViewer   bool        `json:"has_blocked_viewer"`
				ID                 string      `json:"id"`
				IsPrivate          bool        `json:"is_private"`
				IsUnpublished      bool        `json:"is_unpublished"`
				IsVerified         bool        `json:"is_verified"`
				ProfilePicURL      string      `json:"profile_pic_url"`
				RequestedByViewer  bool        `json:"requested_by_viewer"`
				RestrictedByViewer interface{} `json:"restricted_by_viewer"`
				Username           string      `json:"username"`
			} `json:"owner"`
			SensitivityFrictionInfo    interface{} `json:"sensitivity_friction_info"`
			Shortcode                  string      `json:"shortcode"`
			TakenAtTimestamp           int64       `json:"taken_at_timestamp"`
			TrackingToken              string      `json:"tracking_token"`
			ViewerCanReshare           bool        `json:"viewer_can_reshare"`
			ViewerHasLiked             bool        `json:"viewer_has_liked"`
			ViewerHasSaved             bool        `json:"viewer_has_saved"`
			ViewerHasSavedToCollection bool        `json:"viewer_has_saved_to_collection"`
			ViewerInPhotoOfYou         bool        `json:"viewer_in_photo_of_you"`
		} `json:"shortcode_media"`
	} `json:"graphql"`
}

// ToCompletedPost returns a formatted post to persist
func (p *RawPost) ToCompletedPost() (CompletedPost, error) {
	scm := p.Graphql.ShortcodeMedia

	caption := ""
	captionEdges := scm.EdgeMediaToCaption.Edges
	if len(captionEdges) > 0 {
		caption = captionEdges[0].Node.Text
	}

	var re = regexp.MustCompile(`#[^#\s,]+`)
	var tags []string
	for _, match := range re.FindAllString(caption, -1) {
		tags = append(tags, match)
	}

	if scm.ID == "" {
		bytes, err := json.MarshalIndent(p, "", "    ")
		if err != nil {
			return CompletedPost{}, fmt.Errorf("failed to unmarshal json for error message: %s", err)
		}
		return CompletedPost{}, fmt.Errorf("shortcode media was missing ID after fomatting as CompletedPost from raw post: %s", string(bytes))
	}

	return CompletedPost{
		Caption:    caption,
		Code:       scm.Shortcode,
		Dimensions: scm.Dimensions,
		DisplayURL: scm.DisplayURL,
		ID:         scm.ID,
		IsVideo:    scm.IsVideo,
		Location:   scm.Location,
		MediaURL:   scm.DisplayURL,
		PostURL:    "https://instagram.com/p/" + scm.Shortcode,
		Tags:       tags,
		Timestamp:  scm.TakenAtTimestamp,
	}, nil
}
