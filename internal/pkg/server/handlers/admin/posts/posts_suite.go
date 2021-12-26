package posts

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
	"github.com/charlieegan3/photos/cms/internal/pkg/models"
	"github.com/charlieegan3/photos/cms/internal/pkg/server/templating"
)

type EndpointsPostsSuite struct {
	suite.Suite
	DB            *sql.DB
	Bucket        *blob.Bucket
	BucketBaseURL string
}

func (s *EndpointsPostsSuite) SetupTest() {
	err := database.Truncate(s.DB, "posts")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "devices")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "locations")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "medias")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "tags")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "taggings")
	require.NoError(s.T(), err)
}

func (s *EndpointsPostsSuite) TestListPosts() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}
	returnedDevices, err := database.CreateDevices(s.DB, devices)
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
	returnedMedias, err := database.CreateMedias(s.DB, medias)
	require.NoError(s.T(), err)
	locations := []models.Location{
		{
			Name:      "London",
			Latitude:  1.1,
			Longitude: 1.2,
		},
	}

	returnedLocations, err := database.CreateLocations(s.DB, locations)
	require.NoError(s.T(), err)

	posts := []models.Post{
		{
			Description: "Here is a shot I took",
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

	_, err = database.CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc("/admin/posts", BuildIndexHandler(s.DB, templating.BuildPageRenderFunc("http://"))).Methods("GET")

	req, err := http.NewRequest("GET", "/admin/posts", nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	body, err := ioutil.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), string(body), "another photo")
	assert.Contains(s.T(), string(body), "shot")
}

func (s *EndpointsPostsSuite) TestGetPost() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}
	returnedDevices, err := database.CreateDevices(s.DB, devices)
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
	returnedMedias, err := database.CreateMedias(s.DB, medias)
	require.NoError(s.T(), err)
	locations := []models.Location{
		{
			Name:      "London",
			Latitude:  1.1,
			Longitude: 1.2,
		},
	}

	returnedLocations, err := database.CreateLocations(s.DB, locations)
	require.NoError(s.T(), err)

	posts := []models.Post{
		{
			Description: "Here is a shot I took",
			PublishDate: time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}

	persistedPosts, err := database.CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	tags := []models.Tag{
		{Name: "tag_a"},
		{Name: "tag_b"},
		{Name: "tag_c"},
	}
	persistedTags, err := database.CreateTags(s.DB, tags)
	require.NoError(s.T(), err)

	taggings := []models.Tagging{
		{PostID: persistedPosts[0].ID, TagID: persistedTags[0].ID},
		{PostID: persistedPosts[0].ID, TagID: persistedTags[1].ID},
	}
	_, err = database.CreateTaggings(s.DB, taggings)
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc("/admin/posts/{postID}", BuildGetHandler(s.DB, templating.BuildPageRenderFunc("http://"))).Methods("GET")

	req, err := http.NewRequest("GET", fmt.Sprintf("/admin/posts/%d", persistedPosts[0].ID), nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if !assert.Equal(s.T(), http.StatusOK, rr.Code) {
		bodyString, err := ioutil.ReadAll(rr.Body)
		require.NoError(s.T(), err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	body, err := ioutil.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), string(body), "shot I took")
	assert.Contains(s.T(), string(body), "tag_a")
	assert.NotContains(s.T(), string(body), "tag_c")
}

func (s *EndpointsPostsSuite) TestNewPost() {
	router := mux.NewRouter()
	router.HandleFunc("/admin/posts/new", BuildNewHandler(s.DB, templating.BuildPageRenderFunc("http://"))).Methods("GET")

	req, err := http.NewRequest("GET", "/admin/posts/new", nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	body, err := ioutil.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), string(body), "Description")
	assert.Contains(s.T(), string(body), "PublishDate")
	assert.Contains(s.T(), string(body), "LocationID")
	assert.Contains(s.T(), string(body), "MediaID")
}

func (s *EndpointsPostsSuite) TestCreatePost() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}
	returnedDevices, err := database.CreateDevices(s.DB, devices)
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
	returnedMedias, err := database.CreateMedias(s.DB, medias)
	require.NoError(s.T(), err)
	locations := []models.Location{
		{
			Name:      "London",
			Latitude:  1.1,
			Longitude: 1.2,
		},
	}

	returnedLocations, err := database.CreateLocations(s.DB, locations)
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc("/admin/posts", BuildCreateHandler(s.DB, templating.BuildPageRenderFunc("http://"))).Methods("POST")

	form := url.Values{}
	form.Add("Description", "foobar")
	form.Add("IsDraft", "true")
	form.Add("PublishDate", time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC).Format("2006-01-02T15:04"))
	form.Add("LocationID", fmt.Sprintf("%d", returnedLocations[0].ID))
	form.Add("MediaID", fmt.Sprintf("%d", returnedMedias[0].ID))
	form.Add("Tags", "tag_a tagb tag_c")

	// make the request to the handler
	req, err := http.NewRequest(
		"POST",
		"/admin/posts",
		strings.NewReader(form.Encode()),
	)
	require.NoError(s.T(), err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// check that we get a see other response to the right location
	if !assert.Equal(s.T(), http.StatusSeeOther, rr.Code) {
		bodyString, err := ioutil.ReadAll(rr.Body)
		require.NoError(s.T(), err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}
	if !strings.HasPrefix(rr.HeaderMap["Location"][0], "/admin/posts/") {
		s.T().Fatalf("%v doesn't appear to be the correct path", rr.HeaderMap["Location"])
	}

	// check that the database content is also correct
	returnedPosts, err := database.AllPosts(s.DB, true, database.SelectOptions{})
	require.NoError(s.T(), err)

	expectedPosts := td.Slice(
		[]models.Post{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Post{
					Description: "foobar",
					IsDraft:     true,
					PublishDate: time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedPosts, expectedPosts)

	taggings, err := database.FindTaggingsByPostID(s.DB, returnedPosts[0].ID)
	require.NoError(s.T(), err)

	if len(taggings) != 3 {
		s.T().Errorf("expected there to be three taggings")
	}
}

func (s *EndpointsPostsSuite) TestUpdatePost() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}
	returnedDevices, err := database.CreateDevices(s.DB, devices)
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
	returnedMedias, err := database.CreateMedias(s.DB, medias)
	require.NoError(s.T(), err)
	locations := []models.Location{
		{
			Name:      "London",
			Latitude:  1.1,
			Longitude: 1.2,
		},
	}

	returnedLocations, err := database.CreateLocations(s.DB, locations)
	require.NoError(s.T(), err)

	posts := []models.Post{
		{
			Description: "Here is a shot I took",
			PublishDate: time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}

	persistedPosts, err := database.CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	tags := []models.Tag{
		{Name: "tag_a"},
		{Name: "tag_b"},
		{Name: "tag_c"},
	}
	persistedTags, err := database.CreateTags(s.DB, tags)
	require.NoError(s.T(), err)

	taggings := []models.Tagging{
		{PostID: persistedPosts[0].ID, TagID: persistedTags[0].ID},
		{PostID: persistedPosts[0].ID, TagID: persistedTags[1].ID},
	}
	_, err = database.CreateTaggings(s.DB, taggings)
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc("/admin/posts/{postID}", BuildFormHandler(s.DB, templating.BuildPageRenderFunc("http://"))).Methods("POST")

	form := url.Values{}
	form.Add("_method", "PUT")
	form.Add("Description", "foobar")
	form.Add("PublishDate", time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC).Format("2006-01-02T15:04"))
	form.Add("IsDraft", "true")
	form.Add("MediaID", fmt.Sprint(returnedMedias[0].ID))
	form.Add("LocationID", fmt.Sprint(returnedLocations[0].ID))
	form.Add("Tags", " tag_d   \n")

	// make the request to the handler
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/admin/posts/%d", persistedPosts[0].ID),
		strings.NewReader(form.Encode()),
	)
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if !assert.Equal(s.T(), http.StatusSeeOther, rr.Code) {
		bodyString, err := ioutil.ReadAll(rr.Body)
		require.NoError(s.T(), err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	// check that the database content is also correct
	returnedPosts, err := database.AllPosts(s.DB, true, database.SelectOptions{})
	if err != nil {
		s.T().Fatalf("failed to list posts: %s", err)
	}
	expectedPosts := td.Slice(
		[]models.Post{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Post{
					Description: "foobar",
					PublishDate: time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
					MediaID:     returnedMedias[0].ID,
					LocationID:  returnedLocations[0].ID,
					IsDraft:     true,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedPosts, expectedPosts)

	persistedTaggings, err := database.FindTaggingsByPostID(s.DB, returnedPosts[0].ID)
	require.NoError(s.T(), err)

	if len(persistedTaggings) != 1 {
		s.T().Errorf("expected there to be one tagging")
	}

	tagD, err := database.FindTagsByName(s.DB, []string{"tag_d"})
	require.NoError(s.T(), err)

	require.Equal(s.T(), tagD[0].ID, persistedTaggings[0].TagID)
}

func (s *EndpointsPostsSuite) TestDeletePost() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}
	returnedDevices, err := database.CreateDevices(s.DB, devices)
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
	returnedMedias, err := database.CreateMedias(s.DB, medias)
	require.NoError(s.T(), err)
	locations := []models.Location{
		{
			Name:      "London",
			Latitude:  1.1,
			Longitude: 1.2,
		},
	}

	returnedLocations, err := database.CreateLocations(s.DB, locations)
	require.NoError(s.T(), err)

	posts := []models.Post{
		{
			Description: "Here is a shot I took",
			PublishDate: time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}

	persistedPosts, err := database.CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc(
		"/admin/posts/{postID}",
		BuildFormHandler(s.DB, templating.BuildPageRenderFunc("http://")),
	).Methods("POST")

	form := url.Values{}
	form.Add("_method", "DELETE")

	// make the request to the handler
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/admin/posts/%d", persistedPosts[0].ID),
		strings.NewReader(form.Encode()),
	)
	require.NoError(s.T(), err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusSeeOther, rr.Code)

	// check that the database content is also correct
	returnedPosts, err := database.AllPosts(s.DB, true, database.SelectOptions{})
	if err != nil {
		s.T().Fatalf("failed to list posts: %s", err)
	}

	expectedPosts := []models.Post{}
	td.Cmp(s.T(), returnedPosts, expectedPosts)
}
