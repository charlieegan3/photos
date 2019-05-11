package types

import "testing"

func TestFullID(t *testing.T) {
	post := Post{
		Timestamp: 1000,
		ID:        "xxxx",
	}

	if post.FullID() != "1970-01-01-xxxx" {
		t.Errorf("Incorrect full ID %v", post.FullID())
	}
}
