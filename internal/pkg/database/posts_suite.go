package database

import (
	"database/sql"
	"time"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/charlieegan3/photos/cms/internal/pkg/models"
)

// PostsSuite is a number of tests to define the database integration for
// storing posts
type PostsSuite struct {
	suite.Suite
	DB *sql.DB
}

func (s *PostsSuite) SetupTest() {
	err := Truncate(s.DB, "posts")
	require.NoError(s.T(), err)

	err = Truncate(s.DB, "medias")
	require.NoError(s.T(), err)

	err = Truncate(s.DB, "locations")
	require.NoError(s.T(), err)

	err = Truncate(s.DB, "devices")
	require.NoError(s.T(), err)
}

func (s *PostsSuite) TestCreatePosts() {
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
			Description:   "Here is a photo I took",
			InstagramCode: "abc",
			PublishDate:   time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
			MediaID:       returnedMedias[0].ID,
			LocationID:    returnedLocations[0].ID,
		},
	}

	returnedPosts, err := CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	expectedResult := td.Slice(
		[]models.Post{},
		td.ArrayEntries{
			0: td.SStruct(
				posts[0],
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedPosts, expectedResult)
}

func (s *PostsSuite) TestFindPostsByMediaID() {
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
	if err != nil {
		s.T().Fatalf("failed to create locations: %s", err)
	}

	posts := []models.Post{
		{
			Description:   "Here is a photo I took",
			InstagramCode: "foo",
			PublishDate:   time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
			MediaID:       returnedMedias[0].ID,
			LocationID:    returnedLocations[0].ID,
		},
	}

	returnedPosts, err := CreatePosts(s.DB, posts)
	if err != nil {
		s.T().Fatalf("failed to create posts: %s", err)
	}

	posts[0].ID = returnedPosts[0].ID

	returnedPosts, err = FindPostsByMediaID(s.DB, posts[0].MediaID)
	if err != nil {
		s.T().Fatalf("failed get posts: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Post{},
		td.ArrayEntries{
			0: td.SStruct(
				posts[0],
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedPosts, expectedResult)
}

func (s *PostsSuite) TestFindPostsByInstagramCode() {
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
	if err != nil {
		s.T().Fatalf("failed to create locations: %s", err)
	}

	posts := []models.Post{
		{
			Description:   "Here is a photo I took",
			InstagramCode: "foo",
			PublishDate:   time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
			MediaID:       returnedMedias[0].ID,
			LocationID:    returnedLocations[0].ID,
		},
		{
			Description:   "Here is another photo I took, same but diff",
			InstagramCode: "bar",
			PublishDate:   time.Date(2021, time.November, 25, 19, 56, 0, 0, time.UTC),
			MediaID:       returnedMedias[0].ID,
			LocationID:    returnedLocations[0].ID,
		},
	}

	returnedPosts, err := CreatePosts(s.DB, posts)
	if err != nil {
		s.T().Fatalf("failed to create posts: %s", err)
	}

	posts[0].ID = returnedPosts[0].ID

	returnedPosts, err = FindPostsByInstagramCode(s.DB, posts[0].InstagramCode)
	if err != nil {
		s.T().Fatalf("failed get posts: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Post{},
		td.ArrayEntries{
			0: td.SStruct(
				posts[0],
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedPosts, expectedResult)
}

func (s *PostsSuite) TestFindPostsByID() {
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
	if err != nil {
		s.T().Fatalf("failed to create locations: %s", err)
	}

	posts := []models.Post{
		{
			Description: "Here is a photo I took",
			PublishDate: time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
		{
			Description: "Here is another photo I took, same but diff",
			PublishDate: time.Date(2021, time.November, 25, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}

	returnedPosts, err := CreatePosts(s.DB, posts)
	if err != nil {
		s.T().Fatalf("failed to create posts: %s", err)
	}

	posts[0].ID = returnedPosts[0].ID

	returnedPosts, err = FindPostsByID(s.DB, []int{posts[0].ID})
	if err != nil {
		s.T().Fatalf("failed get posts: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Post{},
		td.ArrayEntries{
			0: td.SStruct(
				posts[0],
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedPosts, expectedResult)
}

func (s *PostsSuite) TestFindNextPost() {
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
	if err != nil {
		s.T().Fatalf("failed to create locations: %s", err)
	}

	posts := []models.Post{
		{
			Description: "post 1",
			PublishDate: time.Date(2021, time.October, 25, 19, 56, 0, 0, time.Local),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
		{
			Description: "post 2",
			PublishDate: time.Date(2021, time.November, 24, 19, 56, 0, 0, time.Local),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
		{
			Description: "post 3",
			PublishDate: time.Date(2021, time.December, 25, 19, 56, 0, 0, time.Local),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}

	returnedPosts, err := CreatePosts(s.DB, posts)
	if err != nil {
		s.T().Fatalf("failed to create posts: %s", err)
	}

	posts[0].ID = returnedPosts[0].ID

	nextPosts, err := FindNextPost(s.DB, posts[1], false)
	if err != nil {
		s.T().Fatalf("failed get next: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Post{},
		td.ArrayEntries{
			0: td.SStruct(
				returnedPosts[2],
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), nextPosts, expectedResult)

	nextPosts, err = FindNextPost(s.DB, posts[1], true)
	if err != nil {
		s.T().Fatalf("failed get prev: %s", err)
	}

	expectedResult = td.Slice(
		[]models.Post{},
		td.ArrayEntries{
			0: td.SStruct(
				returnedPosts[0],
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), nextPosts, expectedResult)
}

func (s *PostsSuite) TestCountPosts() {
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
		{
			Description: "Here is another photo I took, same but diff",
			PublishDate: time.Date(2021, time.November, 24, 18, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
			IsDraft:     true,
		},
	}

	_, err = CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	count, err := CountPosts(s.DB, false, SelectOptions{})
	require.NoError(s.T(), err)

	td.Cmp(s.T(), count, int64(1))
}

func (s *PostsSuite) TestAllPosts() {
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
		{
			Description: "Here is another photo I took, same but diff",
			PublishDate: time.Date(2021, time.November, 24, 18, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}

	_, err = CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	returnedPosts, err := AllPosts(s.DB, true, SelectOptions{})
	require.NoError(s.T(), err)

	expectedResult := td.Slice(
		[]models.Post{},
		td.ArrayEntries{
			0: td.SStruct(
				posts[0],
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				posts[1],
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedPosts, expectedResult)
}

func (s *PostsSuite) TestDeletePosts() {
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
		{
			Description: "Here is another photo I took, same but diff",
			PublishDate: time.Date(2021, time.November, 25, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}

	returnedPosts, err := CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	postToDelete := returnedPosts[0]

	err = DeletePosts(s.DB, []models.Post{postToDelete})
	require.NoError(s.T(), err)

	allPosts, err := AllPosts(s.DB, true, SelectOptions{})
	require.NoError(s.T(), err)

	expectedResult := td.Slice(
		[]models.Post{},
		td.ArrayEntries{
			0: td.SStruct(
				posts[1],
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), allPosts, expectedResult)
}

func (s *PostsSuite) TestUpdatePosts() {
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
		{
			Description: "Here is another photo I took, same but diff",
			PublishDate: time.Date(2021, time.November, 25, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}

	createdPosts, err := CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	posts[0].ID = createdPosts[0].ID // needed to match up the update
	posts[1].ID = createdPosts[1].ID // needed to match up the update

	posts[0].Description = "foobar"

	updatedPosts, err := UpdatePosts(s.DB, posts)
	if err != nil {
		s.T().Fatalf("failed to update posts: %s", err)
	}

	expectedPosts := td.Slice(
		[]models.Post{},
		td.ArrayEntries{
			0: td.SStruct(
				posts[0],
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				posts[1],
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), updatedPosts, expectedPosts)
}

func (s *PostsSuite) TestSetPostTags() {
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

	persistedPosts, err := CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	tags := []models.Tag{
		{Name: "tag_a"},
		{Name: "tag_b"},
		{Name: "tag_c"},
	}
	persistedTags, err := CreateTags(s.DB, tags)
	require.NoError(s.T(), err)

	taggings := []models.Tagging{
		{PostID: persistedPosts[0].ID, TagID: persistedTags[0].ID},
		{PostID: persistedPosts[0].ID, TagID: persistedTags[1].ID},
	}
	_, err = CreateTaggings(s.DB, taggings)
	require.NoError(s.T(), err)

	// update to a and c
	err = SetPostTags(s.DB, persistedPosts[0], []string{"tag_a", "tag_c"})
	require.NoError(s.T(), err)

	postTaggings, err := FindTaggingsByPostID(s.DB, persistedPosts[0].ID)
	require.NoError(s.T(), err)

	var tagIDs []int
	for _, v := range postTaggings {
		tagIDs = append(tagIDs, v.TagID)
	}

	postTags, err := FindTagsByID(s.DB, tagIDs)
	require.NoError(s.T(), err)

	expectedResult := td.Slice(
		[]models.Tag{},
		td.ArrayEntries{
			0: td.SStruct(
				tags[0],
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				tags[2],
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), postTags, expectedResult)

	// update to none
	err = SetPostTags(s.DB, persistedPosts[0], []string{})
	require.NoError(s.T(), err)

	postTaggings, err = FindTaggingsByPostID(s.DB, persistedPosts[0].ID)
	require.NoError(s.T(), err)

	tagIDs = []int{}
	for _, v := range postTaggings {
		tagIDs = append(tagIDs, v.TagID)
	}

	postTags, err = FindTagsByID(s.DB, tagIDs)
	require.NoError(s.T(), err)

	expectedResult = td.Slice([]models.Tag{}, td.ArrayEntries{})

	td.Cmp(s.T(), postTags, expectedResult)
}

func (s *PostsSuite) TestPostsInDateRange() {
	devices := []models.Device{{Name: "Example Device"}}
	returnedDevices, err := CreateDevices(s.DB, devices)
	require.NoError(s.T(), err)

	medias := []models.Media{{DeviceID: returnedDevices[0].ID}}
	returnedMedias, err := CreateMedias(s.DB, medias)
	require.NoError(s.T(), err)

	locations := []models.Location{{Name: "London", Latitude: 1.1, Longitude: 1.2}}
	returnedLocations, err := CreateLocations(s.DB, locations)
	require.NoError(s.T(), err)

	posts := []models.Post{
		{
			Description: "older post",
			PublishDate: time.Date(2021, time.October, 24, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
		{
			Description: "post in range",
			PublishDate: time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
		{
			Description: "future post",
			PublishDate: time.Date(2021, time.December, 25, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}

	_, err = CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	returnedPosts, err := PostsInDateRange(
		s.DB,
		time.Date(2021, time.November, 1, 0, 0, 0, 0, time.Local),
		time.Date(2021, time.November, 30, 0, 0, 0, 0, time.Local),
	)
	require.NoError(s.T(), err)

	expectedResult := td.Slice(
		[]models.Post{},
		td.ArrayEntries{
			0: td.SStruct(
				posts[1],
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedPosts, expectedResult)
}
