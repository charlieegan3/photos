package tags

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
	"github.com/charlieegan3/photos/cms/internal/pkg/models"
	"github.com/charlieegan3/photos/cms/internal/pkg/server/templating"
	"github.com/gorilla/mux"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"
)

type EndpointsTagsSuite struct {
	suite.Suite
	DB            *sql.DB
	Bucket        *blob.Bucket
	BucketBaseURL string
}

func (s *EndpointsTagsSuite) SetupTest() {
	err := database.Truncate(s.DB, "tags")
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
	router.HandleFunc("/admin/tags", BuildIndexHandler(s.DB, templating.BuildPageRenderFunc("http://"))).Methods("GET")

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
	router.HandleFunc("/admin/tags/{tagName}", BuildGetHandler(s.DB, templating.BuildPageRenderFunc("http://"))).Methods("GET")

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
	router.HandleFunc("/admin/tags/new", BuildNewHandler(templating.BuildPageRenderFunc("http://"))).Methods("GET")

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
	router.HandleFunc("/admin/tags", BuildCreateHandler(s.DB, templating.BuildPageRenderFunc("http://"))).Methods("POST")

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
	returnedTags, err := database.AllTags(s.DB)
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
	router.HandleFunc("/admin/tags/{tagName}", BuildFormHandler(s.DB, templating.BuildPageRenderFunc("http://"))).Methods("POST")

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

	require.Equal(s.T(), http.StatusSeeOther, rr.Code)

	// check that the database content is also correct
	returnedTags, err := database.AllTags(s.DB)
	if err != nil {
		s.T().Fatalf("failed to list tags: %s", err)
	}
	expectedTags := td.Slice(
		[]models.Tag{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Tag{
					ID:     persistedTags[0].ID,
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
		BuildFormHandler(s.DB, templating.BuildPageRenderFunc("http://")),
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
	returnedTags, err := database.AllTags(s.DB)
	if err != nil {
		s.T().Fatalf("failed to list tags: %s", err)
	}

	expectedTags := []models.Tag{}
	td.Cmp(s.T(), returnedTags, expectedTags)
}
