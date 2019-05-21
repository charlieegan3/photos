package types

import (
	"math"
)

// Location represents the Instgram location entity associated with a Post
type Location struct {
	ID   string  `json:"id"`
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
	Name string  `json:"name"`
	Slug string  `json:"slug"`

	Posts  []Post     `json:"posts"`
	Nearby []Location `json:"locations"`
}

// SetPosts takes a list of posts and stores the ones with matching location on
// the object
func (l *Location) SetPosts(posts []Post) {
	for _, v := range posts {
		if v.Location.ID == l.ID {
			l.Posts = append(l.Posts, v)
		}
	}
}

// SetNearby takes a list of locations and selects the ones that are within the
// distance to the location
func (l *Location) SetNearby(locations []Location, kilometres int) {
	for _, v := range locations {
		if l.isNearby(v, kilometres) {
			l.Nearby = append(l.Nearby, v)
		}
	}
}

func (l *Location) isNearby(location Location, kilometres int) bool {
	if l.equals(location) {
		return false
	}

	d := l.distance(location)
	return d < float64(kilometres)
}

func (l *Location) equals(location Location) bool {
	return l.Lat == location.Lat && l.Long == location.Long
}

func (l *Location) distance(location Location) float64 {
	radlat1 := float64(math.Pi * l.Lat / 180)
	radlat2 := float64(math.Pi * location.Lat / 180)
	radtheta := float64(math.Pi * float64(l.Long-location.Long) / 180)

	distance := math.Sin(radlat1)*math.Sin(radlat2) + math.Cos(radlat1)*math.Cos(radlat2)*math.Cos(radtheta)

	if distance > 1 {
		distance = 1
	}

	distance = math.Acos(distance)
	distance = distance * 180 / math.Pi

	return distance * 60 * 1.1515 * 1.609344
}
