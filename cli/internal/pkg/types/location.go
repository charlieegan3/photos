package types

import (
	"math"
	"sort"
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
	var nearby []struct {
		Distance float64
		Location Location
	}

	for _, v := range locations {
		if l.equals(v) {
			continue
		}
		distance := l.distance(v)
		if distance < float64(kilometres) {
			item := struct {
				Distance float64
				Location Location
			}{
				Distance: distance, Location: v,
			}
			nearby = append(nearby, item)
		}
	}

	sort.SliceStable(nearby, func(i, j int) bool {
		return nearby[j].Distance > nearby[i].Distance
	})

	for i, v := range nearby {
		if i > 5 {
			continue
		}
		l.Nearby = append(l.Nearby, v.Location)
	}
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
