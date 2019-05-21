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

func TestNearbyLocations(t *testing.T) {
	location1 := Location{
		ID:  "l1",
		Lat: 57.5310428, Long: -4.4104024,
	}
	location2 := Location{
		ID:  "l2",
		Lat: 57.5327879, Long: -4.4013311,
	}
	location3 := Location{
		ID:  "l3",
		Lat: 57.4803603, Long: -4.2212256,
	}
	locations := []Location{location1, location2, location3}

	location1.SetNearby(locations, 3)

	if len(location1.Nearby) != 1 {
		t.Error("Unexpected number of nearby locations")
	}

	if location1.Nearby[0].ID != "l2" {
		t.Error("Unexpected nearby location")
	}
}
