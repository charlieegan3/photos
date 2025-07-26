package posts

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/suite"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"

	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/charlieegan3/photos/internal/pkg/models"
	"github.com/charlieegan3/photos/internal/pkg/server/templating"
)

type EndpointsPostsSuite struct {
	suite.Suite

	DB            *sql.DB
	Bucket        *blob.Bucket
	BucketBaseURL string
}

func (s *EndpointsPostsSuite) SetupTest() {
	err := database.Truncate(s.T().Context(), s.DB, "photos.posts")
	s.Require().NoError(err)
	err = database.Truncate(s.T().Context(), s.DB, "photos.devices")
	s.Require().NoError(err)
	err = database.Truncate(s.T().Context(), s.DB, "photos.locations")
	s.Require().NoError(err)
	err = database.Truncate(s.T().Context(), s.DB, "photos.medias")
	s.Require().NoError(err)
	err = database.Truncate(s.T().Context(), s.DB, "photos.tags")
	s.Require().NoError(err)
	err = database.Truncate(s.T().Context(), s.DB, "photos.taggings")
	s.Require().NoError(err)
}

func (s *EndpointsPostsSuite) TestListPosts() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}
	returnedDevices, err := database.CreateDevices(s.T().Context(), s.DB, devices)
	s.Require().NoError(err)

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
	returnedMedias, err := database.CreateMedias(s.T().Context(), s.DB, medias)
	s.Require().NoError(err)
	locations := []models.Location{
		{
			Name:      "London",
			Latitude:  1.1,
			Longitude: 1.2,
		},
	}

	returnedLocations, err := database.CreateLocations(s.T().Context(), s.DB, locations)
	s.Require().NoError(err)

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

	_, err = database.CreatePosts(s.T().Context(), s.DB, posts)
	s.Require().NoError(err)

	router := mux.NewRouter()
	router.HandleFunc("/admin/posts",
		BuildIndexHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(s.T().Context(), http.MethodGet, "/admin/posts", nil)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusOK, rr.Code)

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains(string(body), "another photo")
	s.Contains(string(body), "shot")
}

func (s *EndpointsPostsSuite) TestGetPost() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}
	returnedDevices, err := database.CreateDevices(s.T().Context(), s.DB, devices)
	s.Require().NoError(err)

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
	returnedMedias, err := database.CreateMedias(s.T().Context(), s.DB, medias)
	s.Require().NoError(err)
	locations := []models.Location{
		{
			Name:      "London",
			Latitude:  1.1,
			Longitude: 1.2,
		},
	}

	returnedLocations, err := database.CreateLocations(s.T().Context(), s.DB, locations)
	s.Require().NoError(err)

	posts := []models.Post{
		{
			Description: "Here is a shot I took",
			PublishDate: time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}

	persistedPosts, err := database.CreatePosts(s.T().Context(), s.DB, posts)
	s.Require().NoError(err)

	tags := []models.Tag{
		{Name: "tag_a"},
		{Name: "tag_b"},
		{Name: "tag_c"},
	}
	persistedTags, err := database.CreateTags(s.T().Context(), s.DB, tags)
	s.Require().NoError(err)

	taggings := []models.Tagging{
		{PostID: persistedPosts[0].ID, TagID: persistedTags[0].ID},
		{PostID: persistedPosts[0].ID, TagID: persistedTags[1].ID},
	}
	_, err = database.CreateTaggings(s.T().Context(), s.DB, taggings)
	s.Require().NoError(err)

	router := mux.NewRouter()
	router.HandleFunc("/admin/posts/{postID}",
		BuildGetHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(
		s.T().Context(), http.MethodGet, fmt.Sprintf("/admin/posts/%d", persistedPosts[0].ID), nil,
	)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if !s.Equal(http.StatusOK, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		s.Require().NoError(err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains(string(body), "shot I took")
	s.Contains(string(body), "tag_a")
	s.NotContains(string(body), "tag_c")
}

func (s *EndpointsPostsSuite) TestNewPost() {
	router := mux.NewRouter()
	router.HandleFunc("/admin/posts/new",
		BuildNewHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(s.T().Context(), http.MethodGet, "/admin/posts/new", nil)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusOK, rr.Code)

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains(string(body), "Description")
	s.Contains(string(body), "PublishDate")
	s.Contains(string(body), "LocationID")
	s.Contains(string(body), "MediaID")
}

func (s *EndpointsPostsSuite) TestCreatePost() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}
	returnedDevices, err := database.CreateDevices(s.T().Context(), s.DB, devices)
	s.Require().NoError(err)

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
	returnedMedias, err := database.CreateMedias(s.T().Context(), s.DB, medias)
	s.Require().NoError(err)
	locations := []models.Location{
		{
			Name:      "London",
			Latitude:  1.1,
			Longitude: 1.2,
		},
	}

	returnedLocations, err := database.CreateLocations(s.T().Context(), s.DB, locations)
	s.Require().NoError(err)

	router := mux.NewRouter()
	router.HandleFunc("/admin/posts",
		BuildCreateHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodPost)

	form := url.Values{}
	form.Add("Description", "foobar")
	form.Add("IsDraft", "true")
	form.Add("PublishDate", time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC).Format("2006-01-02T15:04"))
	form.Add("LocationID", strconv.Itoa(returnedLocations[0].ID))
	form.Add("MediaID", strconv.Itoa(returnedMedias[0].ID))
	form.Add("Tags", "tag_a tagb tag_c")

	// make the request to the handler
	req, err := http.NewRequestWithContext(
		s.T().Context(),
		http.MethodPost,
		"/admin/posts",
		strings.NewReader(form.Encode()),
	)
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// check that we get a see other response to the right location
	if !s.Equal(http.StatusSeeOther, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		s.Require().NoError(err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}
	if !strings.HasPrefix(rr.Result().Header["Location"][0], "/admin/posts/") {
		s.T().Fatalf("%v doesn't appear to be the correct path", rr.Result().Header["Location"])
	}

	// check that the database content is also correct
	returnedPosts, err := database.AllPosts(s.DB, true, database.SelectOptions{})
	s.Require().NoError(err)

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
	s.Require().NoError(err)

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
	returnedDevices, err := database.CreateDevices(s.T().Context(), s.DB, devices)
	s.Require().NoError(err)

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
	returnedMedias, err := database.CreateMedias(s.T().Context(), s.DB, medias)
	s.Require().NoError(err)
	locations := []models.Location{
		{
			Name:      "London",
			Latitude:  1.1,
			Longitude: 1.2,
		},
	}

	returnedLocations, err := database.CreateLocations(s.T().Context(), s.DB, locations)
	s.Require().NoError(err)

	posts := []models.Post{
		{
			Description: "Here is a shot I took",
			PublishDate: time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}

	persistedPosts, err := database.CreatePosts(s.T().Context(), s.DB, posts)
	s.Require().NoError(err)

	tags := []models.Tag{
		{Name: "tag_a"},
		{Name: "tag_b"},
		{Name: "tag_c"},
	}
	persistedTags, err := database.CreateTags(s.T().Context(), s.DB, tags)
	s.Require().NoError(err)

	taggings := []models.Tagging{
		{PostID: persistedPosts[0].ID, TagID: persistedTags[0].ID},
		{PostID: persistedPosts[0].ID, TagID: persistedTags[1].ID},
	}
	_, err = database.CreateTaggings(s.T().Context(), s.DB, taggings)
	s.Require().NoError(err)

	router := mux.NewRouter()
	router.HandleFunc("/admin/posts/{postID}",
		BuildFormHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodPost)

	form := url.Values{}
	form.Add("_method", http.MethodPut)
	form.Add("Description", "foobar")
	form.Add("PublishDate", time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC).Format("2006-01-02T15:04"))
	form.Add("IsDraft", "true")
	form.Add("MediaID", strconv.Itoa(returnedMedias[0].ID))
	form.Add("LocationID", strconv.Itoa(returnedLocations[0].ID))
	form.Add("Tags", " tag_d   \n")

	// make the request to the handler
	req, err := http.NewRequestWithContext(
		s.T().Context(),
		http.MethodPost,
		fmt.Sprintf("/admin/posts/%d", persistedPosts[0].ID),
		strings.NewReader(form.Encode()),
	)
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if !s.Equal(http.StatusSeeOther, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		s.Require().NoError(err)
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
	s.Require().NoError(err)

	if len(persistedTaggings) != 1 {
		s.T().Errorf("expected there to be one tagging")
	}

	tagD, err := database.FindTagsByName(s.T().Context(), s.DB, []string{"tag_d"})
	s.Require().NoError(err)

	s.Require().Equal(tagD[0].ID, persistedTaggings[0].TagID)
}

func (s *EndpointsPostsSuite) TestDeletePost() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}
	returnedDevices, err := database.CreateDevices(s.T().Context(), s.DB, devices)
	s.Require().NoError(err)

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
	returnedMedias, err := database.CreateMedias(s.T().Context(), s.DB, medias)
	s.Require().NoError(err)
	locations := []models.Location{
		{
			Name:      "London",
			Latitude:  1.1,
			Longitude: 1.2,
		},
	}

	returnedLocations, err := database.CreateLocations(s.T().Context(), s.DB, locations)
	s.Require().NoError(err)

	posts := []models.Post{
		{
			Description: "Here is a shot I took",
			PublishDate: time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}

	persistedPosts, err := database.CreatePosts(s.T().Context(), s.DB, posts)
	s.Require().NoError(err)

	router := mux.NewRouter()
	router.HandleFunc(
		"/admin/posts/{postID}",
		BuildFormHandler(s.DB, templating.BuildPageRenderFunc(true, "")),
	).Methods(http.MethodPost)

	form := url.Values{}
	form.Add("_method", http.MethodDelete)

	// make the request to the handler
	req, err := http.NewRequestWithContext(
		s.T().Context(),
		http.MethodPost,
		fmt.Sprintf("/admin/posts/%d", persistedPosts[0].ID),
		strings.NewReader(form.Encode()),
	)
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusSeeOther, rr.Code)

	// check that the database content is also correct
	returnedPosts, err := database.AllPosts(s.DB, true, database.SelectOptions{})
	if err != nil {
		s.T().Fatalf("failed to list posts: %s", err)
	}

	expectedPosts := []models.Post{}
	td.Cmp(s.T(), returnedPosts, expectedPosts)
}
