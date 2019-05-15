package types

// Location represents the Instgram location entity associated with a Post
type Location struct {
	ID   string  `json:"id"`
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
	Name string  `json:"name"`
	Slug string  `json:"slug"`
}
