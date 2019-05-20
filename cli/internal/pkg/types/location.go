package types

// Location represents the Instgram location entity associated with a Post
type Location struct {
	ID    string  `json:"id"`
	Lat   float64 `json:"lat"`
	Long  float64 `json:"long"`
	Name  string  `json:"name"`
	Slug  string  `json:"slug"`
	Posts []Post  `json:"posts"`
}

// SetPosts takes a list of posts and stores the ones with matching location on the object
func (p *Location) SetPosts(posts []Post) {
	for _, v := range posts {
		if v.Location.ID == p.ID {
			p.Posts = append(p.Posts, v)
		}
	}
}
