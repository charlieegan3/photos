package database

import (
	"database/sql"
	"time"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/charlieegan3/photos/internal/pkg/models"
)

// TaggingsSuite is a number of tests to define the database integration for
// storing taggings
type TaggingsSuite struct {
	suite.Suite
	DB *sql.DB
}

func (s *TaggingsSuite) SetupTest() {
	err := Truncate(s.DB, "taggings")
	require.NoError(s.T(), err)

	err = Truncate(s.DB, "medias")
	require.NoError(s.T(), err)

	err = Truncate(s.DB, "locations")
	require.NoError(s.T(), err)

	err = Truncate(s.DB, "devices")
	require.NoError(s.T(), err)

	err = Truncate(s.DB, "tags")
	require.NoError(s.T(), err)
}

func (s *TaggingsSuite) TestCreateTaggings() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}
	returnedDevices, err := CreateDevices(s.DB, devices)
	require.NoError(s.T(), err)

	medias := []models.Media{
		{
			DeviceID: returnedDevices[0].ID,

			Make:  "FujiFilm",
			Model: "X100F",

			TakenAt: time.Date(2021, time.November, 23, 19, 56, 0, 0, time.UTC),

			FNumber:  2.0,
			ISOSpeed: 100,

			Latitude:  51.1,
			Longitude: 52.2,
			Altitude:  100.0,
		},
	}
	returnedMedias, err := CreateMedias(s.DB, medias)
	require.NoError(s.T(), err)
	locations := []models.Location{
		{
			Name:      "London",
			Latitude:  1.1,
			Longitude: 1.2,
		},
	}

	returnedLocations, err := CreateLocations(s.DB, locations)
	require.NoError(s.T(), err)

	posts := []models.Post{
		{
			Description: "Here is a photo I took",
			PublishDate: time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}
	returnedPosts, err := CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	tags := []models.Tag{
		{
			Name: "nofilter",
		},
	}
	returnedTags, err := CreateTags(s.DB, tags)
	require.NoError(s.T(), err)

	taggings := []models.Tagging{
		{
			PostID: returnedPosts[0].ID,
			TagID:  returnedTags[0].ID,
		},
	}

	returnedTaggings, err := CreateTaggings(s.DB, taggings)
	require.NoError(s.T(), err)

	expectedResult := td.Slice(
		[]models.Tagging{},
		td.ArrayEntries{
			0: td.SStruct(
				taggings[0],
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedTaggings, expectedResult)
}

func (s *TaggingsSuite) TestFindOrCreateTaggings() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}
	returnedDevices, err := CreateDevices(s.DB, devices)
	require.NoError(s.T(), err)

	medias := []models.Media{
		{
			DeviceID: returnedDevices[0].ID,

			Make:  "FujiFilm",
			Model: "X100F",

			TakenAt: time.Date(2021, time.November, 23, 19, 56, 0, 0, time.UTC),

			FNumber:  2.0,
			ISOSpeed: 100,

			Latitude:  51.1,
			Longitude: 52.2,
			Altitude:  100.0,
		},
	}
	returnedMedias, err := CreateMedias(s.DB, medias)
	require.NoError(s.T(), err)
	locations := []models.Location{
		{
			Name:      "London",
			Latitude:  1.1,
			Longitude: 1.2,
		},
	}

	returnedLocations, err := CreateLocations(s.DB, locations)
	require.NoError(s.T(), err)

	posts := []models.Post{
		{
			Description: "Here is a photo I took",
			PublishDate: time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}
	returnedPosts, err := CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	tags := []models.Tag{
		{
			Name: "nofilter",
		},
	}
	returnedTags, err := CreateTags(s.DB, tags)
	require.NoError(s.T(), err)

	taggings := []models.Tagging{
		{
			PostID: returnedPosts[0].ID,
			TagID:  returnedTags[0].ID,
		},
	}

	returnedTaggings, err := FindOrCreateTaggings(s.DB, taggings)
	require.NoError(s.T(), err)

	expectedResult := td.Slice(
		[]models.Tagging{},
		td.ArrayEntries{
			0: td.SStruct(
				taggings[0],
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedTaggings, expectedResult)

	// create them again
	returnedTaggings2, err := FindOrCreateTaggings(s.DB, taggings)
	require.NoError(s.T(), err)

	td.Cmp(s.T(), returnedTaggings, returnedTaggings2)
}

func (s *TaggingsSuite) TestFindTaggingsByPostID() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}
	returnedDevices, err := CreateDevices(s.DB, devices)
	require.NoError(s.T(), err)

	medias := []models.Media{
		{
			DeviceID: returnedDevices[0].ID,

			Make:  "FujiFilm",
			Model: "X100F",

			TakenAt: time.Date(2021, time.November, 23, 19, 56, 0, 0, time.UTC),

			FNumber:  2.0,
			ISOSpeed: 100,

			Latitude:  51.1,
			Longitude: 52.2,
			Altitude:  100.0,
		},
	}
	returnedMedias, err := CreateMedias(s.DB, medias)
	require.NoError(s.T(), err)
	locations := []models.Location{
		{
			Name:      "London",
			Latitude:  1.1,
			Longitude: 1.2,
		},
	}

	returnedLocations, err := CreateLocations(s.DB, locations)
	require.NoError(s.T(), err)

	posts := []models.Post{
		{
			Description: "Here is a photo I took",
			PublishDate: time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}
	returnedPosts, err := CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	tags := []models.Tag{
		{
			Name: "nofilter",
		},
	}
	returnedTags, err := CreateTags(s.DB, tags)
	require.NoError(s.T(), err)

	taggings := []models.Tagging{
		{
			PostID: returnedPosts[0].ID,
			TagID:  returnedTags[0].ID,
		},
	}

	returnedTaggings, err := CreateTaggings(s.DB, taggings)
	require.NoError(s.T(), err)

	returnedTaggings, err = FindTaggingsByPostID(s.DB, returnedPosts[0].ID)
	require.NoError(s.T(), err)

	expectedResult := td.Slice(
		[]models.Tagging{},
		td.ArrayEntries{
			0: td.SStruct(
				taggings[0],
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedTaggings, expectedResult)
}

func (s *TaggingsSuite) TestDeleteTaggings() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}
	returnedDevices, err := CreateDevices(s.DB, devices)
	require.NoError(s.T(), err)

	medias := []models.Media{
		{
			DeviceID: returnedDevices[0].ID,

			Make:  "FujiFilm",
			Model: "X100F",

			TakenAt: time.Date(2021, time.November, 23, 19, 56, 0, 0, time.UTC),

			FNumber:  2.0,
			ISOSpeed: 100,

			Latitude:  51.1,
			Longitude: 52.2,
			Altitude:  100.0,
		},
	}
	returnedMedias, err := CreateMedias(s.DB, medias)
	require.NoError(s.T(), err)
	locations := []models.Location{
		{
			Name:      "London",
			Latitude:  1.1,
			Longitude: 1.2,
		},
	}

	returnedLocations, err := CreateLocations(s.DB, locations)
	require.NoError(s.T(), err)

	posts := []models.Post{
		{
			Description: "Here is a photo I took",
			PublishDate: time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}
	returnedPosts, err := CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	tags := []models.Tag{
		{
			Name: "nofilter",
		},
	}
	returnedTags, err := CreateTags(s.DB, tags)
	require.NoError(s.T(), err)

	taggings := []models.Tagging{
		{
			PostID: returnedPosts[0].ID,
			TagID:  returnedTags[0].ID,
		},
	}

	returnedTaggings, err := CreateTaggings(s.DB, taggings)
	require.NoError(s.T(), err)

	taggingToDelete := returnedTaggings[0]

	err = DeleteTaggings(s.DB, []models.Tagging{taggingToDelete})
	require.NoError(s.T(), err)

	postTaggings, err := FindTaggingsByPostID(s.DB, returnedPosts[0].ID)
	require.NoError(s.T(), err)

	if len(postTaggings) > 0 {
		s.T().Fatalf("expected there to be no post taggings, but there were some")
	}
}
