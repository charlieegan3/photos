package tags

import (
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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

type EndpointsTagsSuite struct {
	suite.Suite

	DB            *sql.DB
	Bucket        *blob.Bucket
	BucketBaseURL string
}

func (s *EndpointsTagsSuite) SetupTest() {
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

func (s *EndpointsTagsSuite) TestListTags() {
	testData := []models.Tag{
		{
			Name:   "nofilter",
			Hidden: true,
		},
		{
			Name: "x100f",
		},
	}

	_, err := database.CreateTags(s.T().Context(), s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create tags: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/tags",
		BuildIndexHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(s.T().Context(), http.MethodGet, "/admin/tags", nil)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusOK, rr.Code)

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains(string(body), "x100f")
	s.Contains(string(body), "nofilter")
}

func (s *EndpointsTagsSuite) TestGetTag() {
	testData := []models.Tag{
		{
			Name:   "nofilter",
			Hidden: true,
		},
	}

	persistedTags, err := database.CreateTags(s.T().Context(), s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create tags: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/tags/{tagName}",
		BuildGetHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(s.T().Context(), http.MethodGet, "/admin/tags/"+persistedTags[0].Name, nil)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusOK, rr.Code)

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains(string(body), "nofilter")
	s.Contains(string(body), `name="Hidden" type="checkbox" value="true" checked`)
}

func (s *EndpointsTagsSuite) TestNewTag() {
	router := mux.NewRouter()
	router.HandleFunc("/admin/tags/new", BuildNewHandler(templating.BuildPageRenderFunc(true, ""))).Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(s.T().Context(), http.MethodGet, "/admin/tags/new", nil)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusOK, rr.Code)

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains(string(body), "Name")
	s.Contains(string(body), "Hidden")
}

func (s *EndpointsTagsSuite) TestCreateTag() {
	router := mux.NewRouter()
	router.HandleFunc("/admin/tags",
		BuildCreateHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodPost)

	form := url.Values{}
	form.Add("Name", "nofilter")
	form.Add("Hidden", "true")

	// make the request to the handler
	req, err := http.NewRequestWithContext(
		s.T().Context(),
		http.MethodPost,
		"/admin/tags",
		strings.NewReader(form.Encode()),
	)
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// check that we get a see other response to the right location
	s.Require().Equal(http.StatusSeeOther, rr.Code)
	td.Cmp(s.T(), rr.Result().Header["Location"], []string{"/admin/tags/nofilter"})

	// check that the database content is also correct
	returnedTags, err := database.AllTags(s.T().Context(), s.DB, true, database.SelectOptions{})
	s.Require().NoError(err)

	expectedTags := td.Slice(
		[]models.Tag{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Tag{
					Name:   "nofilter",
					Hidden: true,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedTags, expectedTags)
}

func (s *EndpointsTagsSuite) TestUpdateTag() {
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
			Description: "Here is another shot I took",
			PublishDate: time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}
	returnedPosts, err := database.CreatePosts(s.T().Context(), s.DB, posts)
	s.Require().NoError(err)

	testData := []models.Tag{
		{
			Name:   "nofilter",
			Hidden: true,
		},
		{
			Name:   "nofiltered",
			Hidden: false,
		},
		{
			Name:   "tag1",
			Hidden: false,
		},
		{
			Name:   "tag2",
			Hidden: false,
		},
	}

	persistedTags, err := database.CreateTags(s.T().Context(), s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create tags: %s", err)
	}

	err = database.SetPostTags(s.T().Context(), s.DB, returnedPosts[0], []string{"tag1"})
	s.Require().NoError(err)
	err = database.SetPostTags(s.T().Context(), s.DB, returnedPosts[1], []string{"tag2"})
	s.Require().NoError(err)

	router := mux.NewRouter()
	router.HandleFunc("/admin/tags/{tagName}",
		BuildFormHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodPost)

	form := url.Values{}
	form.Add("_method", http.MethodPut)
	form.Add("Name", "nofiltered")
	form.Add("Hidden", "false")

	// make the request to the handler
	req, err := http.NewRequestWithContext(
		s.T().Context(),
		http.MethodPost,
		"/admin/tags/"+persistedTags[0].Name,
		strings.NewReader(form.Encode()),
	)
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	form = url.Values{}
	form.Add("_method", http.MethodPut)
	form.Add("Name", "tag1")
	form.Add("Hidden", "false")

	// make the request to the handler
	req, err = http.NewRequestWithContext(
		s.T().Context(), http.MethodPost, "/admin/tags/tag2", strings.NewReader(form.Encode()),
	)
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if !s.Equal(http.StatusSeeOther, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		s.Require().NoError(err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	// check that the database content is also correct
	returnedTags, err := database.AllTags(s.T().Context(), s.DB, true, database.SelectOptions{
		SortField: "name",
	})
	s.Require().NoError(err)

	expectedTags := td.Slice(
		[]models.Tag{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Tag{
					Name:   "nofiltered",
					Hidden: false,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Tag{
					Name:   "tag1",
					Hidden: false,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedTags, expectedTags)

	taggings, err := database.FindTaggingsByTagID(s.T().Context(), s.DB, returnedTags[1].ID)
	s.Require().NoError(err)

	expectedTaggings := td.Slice(
		[]models.Tagging{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Tagging{
					PostID: posts[0].ID,
					TagID:  returnedTags[1].ID,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Tagging{
					PostID: posts[1].ID,
					TagID:  returnedTags[1].ID,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), taggings, expectedTaggings)
}

func (s *EndpointsTagsSuite) TestDeleteTag() {
	testData := []models.Tag{
		{
			Name: "nofilter",
		},
	}

	persistedTags, err := database.CreateTags(s.T().Context(), s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create tags: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc(
		"/admin/tags/{tagName}",
		BuildFormHandler(s.DB, templating.BuildPageRenderFunc(true, "")),
	).Methods(http.MethodPost)

	form := url.Values{}
	form.Add("_method", http.MethodDelete)

	// make the request to the handler
	req, err := http.NewRequestWithContext(
		s.T().Context(),
		http.MethodPost,
		"/admin/tags/"+persistedTags[0].Name,
		strings.NewReader(form.Encode()),
	)
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusSeeOther, rr.Code)

	// check that the database content is also correct
	returnedTags, err := database.AllTags(s.T().Context(), s.DB, false, database.SelectOptions{})
	if err != nil {
		s.T().Fatalf("failed to list tags: %s", err)
	}

	expectedTags := []models.Tag{}
	td.Cmp(s.T(), returnedTags, expectedTags)
}
