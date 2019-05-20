package types

import "testing"

func TestMatchingPosts(t *testing.T) {
	location1 := struct {
		HasPublicPage bool   `json:"has_public_page"`
		ID            string `json:"id"`
		Name          string `json:"name"`
		Slug          string `json:"slug"`
	}{
		ID: "1234",
	}
	location2 := struct {
		HasPublicPage bool   `json:"has_public_page"`
		ID            string `json:"id"`
		Name          string `json:"name"`
		Slug          string `json:"slug"`
	}{
		ID: "5678",
	}
	post1 := Post{
		ID:       "abc",
		Location: location1,
	}
	post2 := Post{
		ID:       "xyz",
		Location: location2,
	}

	location := Location{
		ID: "1234",
	}

	location.SetPosts([]Post{post1, post2})

	if len(location.Posts) != 1 {
		t.Errorf("Unexpected number of posts %v", len(location.Posts))
	}

	if location.Posts[0].ID != "abc" {
		t.Errorf("Unexpected post %v", location.Posts[0].ID)
	}
}
