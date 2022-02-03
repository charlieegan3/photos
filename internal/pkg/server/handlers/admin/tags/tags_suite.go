package tags

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

type EndpointsTagsSuite struct {
	suite.Suite
	DB            *sql.DB
	Bucket        *blob.Bucket
	BucketBaseURL string
}

func (s *EndpointsTagsSuite) SetupTest() {
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

	_, err := database.CreateTags(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create tags: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/tags", BuildIndexHandler(s.DB, templating.BuildPageRenderFunc())).Methods("GET")

	req, err := http.NewRequest("GET", "/admin/tags", nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	body, err := ioutil.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), string(body), "x100f")
	assert.Contains(s.T(), string(body), "nofilter")
}

func (s *EndpointsTagsSuite) TestGetTag() {
	testData := []models.Tag{
		{
			Name:   "nofilter",
			Hidden: true,
		},
	}

	persistedTags, err := database.CreateTags(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create tags: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/tags/{tagName}", BuildGetHandler(s.DB, templating.BuildPageRenderFunc())).Methods("GET")

	req, err := http.NewRequest("GET", fmt.Sprintf("/admin/tags/%s", persistedTags[0].Name), nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	body, err := ioutil.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), string(body), "nofilter")
	assert.Contains(s.T(), string(body), `name="Hidden" type="checkbox" value="true" checked`)
}

func (s *EndpointsTagsSuite) TestNewTag() {
	router := mux.NewRouter()
	router.HandleFunc("/admin/tags/new", BuildNewHandler(templating.BuildPageRenderFunc())).Methods("GET")

	req, err := http.NewRequest("GET", "/admin/tags/new", nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	body, err := ioutil.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), string(body), "Name")
	assert.Contains(s.T(), string(body), "Hidden")
}

func (s *EndpointsTagsSuite) TestCreateTag() {
	router := mux.NewRouter()
	router.HandleFunc("/admin/tags", BuildCreateHandler(s.DB, templating.BuildPageRenderFunc())).Methods("POST")

	form := url.Values{}
	form.Add("Name", "nofilter")
	form.Add("Hidden", "true")

	// make the request to the handler
	req, err := http.NewRequest(
		"POST",
		"/admin/tags",
		strings.NewReader(form.Encode()),
	)
	require.NoError(s.T(), err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// check that we get a see other response to the right location
	require.Equal(s.T(), http.StatusSeeOther, rr.Code)
	td.Cmp(s.T(), rr.HeaderMap["Location"], []string{"/admin/tags/nofilter"})

	// check that the database content is also correct
	returnedTags, err := database.AllTags(s.DB, true, database.SelectOptions{})
	require.NoError(s.T(), err)

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
			Description: "Here is another shot I took",
			PublishDate: time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}
	returnedPosts, err := database.CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	err = database.SetPostTags(s.DB, returnedPosts[0], []string{"tag1"})
	require.NoError(s.T(), err)
	err = database.SetPostTags(s.DB, returnedPosts[1], []string{"tag2"})
	require.NoError(s.T(), err)

	testData := []models.Tag{
		{
			Name:   "nofilter",
			Hidden: true,
		},
		{
			Name:   "nofiltered",
			Hidden: false,
		},
	}

	persistedTags, err := database.CreateTags(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create tags: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/tags/{tagName}", BuildFormHandler(s.DB, templating.BuildPageRenderFunc())).Methods("POST")

	form := url.Values{}
	form.Add("_method", "PUT")
	form.Add("Name", "nofiltered")
	form.Add("Hidden", "false")

	// make the request to the handler
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/admin/tags/%s", persistedTags[0].Name),
		strings.NewReader(form.Encode()),
	)
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	form = url.Values{}
	form.Add("_method", "PUT")
	form.Add("Name", "tag1")
	form.Add("Hidden", "false")

	// make the request to the handler
	req, err = http.NewRequest("POST", "/admin/tags/tag2", strings.NewReader(form.Encode()))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if !assert.Equal(s.T(), http.StatusSeeOther, rr.Code) {
		bodyString, err := ioutil.ReadAll(rr.Body)
		require.NoError(s.T(), err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	// check that the database content is also correct
	returnedTags, err := database.AllTags(s.DB, true, database.SelectOptions{})
	if err != nil {
		s.T().Fatalf("failed to list tags: %s", err)
	}
	expectedTags := td.Slice(
		[]models.Tag{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Tag{
					Name:   "tag1",
					Hidden: false,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Tag{
					ID:     persistedTags[1].ID,
					Name:   "nofiltered",
					Hidden: false,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedTags, expectedTags)
}

func (s *EndpointsTagsSuite) TestDeleteTag() {
	testData := []models.Tag{
		{
			Name: "nofilter",
		},
	}

	persistedTags, err := database.CreateTags(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create tags: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc(
		"/admin/tags/{tagName}",
		BuildFormHandler(s.DB, templating.BuildPageRenderFunc()),
	).Methods("POST")

	form := url.Values{}
	form.Add("_method", "DELETE")

	// make the request to the handler
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/admin/tags/%s", persistedTags[0].Name),
		strings.NewReader(form.Encode()),
	)
	require.NoError(s.T(), err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusSeeOther, rr.Code)

	// check that the database content is also correct
	returnedTags, err := database.AllTags(s.DB, false, database.SelectOptions{})
	if err != nil {
		s.T().Fatalf("failed to list tags: %s", err)
	}

	expectedTags := []models.Tag{}
	td.Cmp(s.T(), returnedTags, expectedTags)
}
