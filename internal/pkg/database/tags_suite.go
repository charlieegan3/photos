package database

import (
	"database/sql"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/charlieegan3/photos/cms/internal/pkg/models"
)

// TagsSuite is a number of tests to define the database integration for
// storing tags
type TagsSuite struct {
	suite.Suite
	DB *sql.DB
}

func (s *TagsSuite) SetupTest() {
	err := Truncate(s.DB, "tags")
	if err != nil {
		s.T().Fatalf("failed to truncate table: %s", err)
	}
}

func (s *TagsSuite) TestCreateTags() {
	tags := []models.Tag{
		{
			Name:   "No Filter",
			Hidden: true,
		},
		{
			Name: "shotoniphone",
		},
	}

	returnedTags, err := CreateTags(s.DB, tags)
	if err != nil {
		s.T().Fatalf("failed to create tags: %s", err)
	}

	expectedTags := td.Slice(
		[]models.Tag{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Tag{
					Name:   "no-filter",
					Hidden: true,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Tag{
					Name:   "shotoniphone",
					Hidden: false,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedTags, expectedTags)
}

func (s *TagsSuite) TestFindOrCreateTagsByName() {
	existingTags := []models.Tag{
		{
			Name:   "foobar",
			Hidden: true,
		},
	}

	returnedTags, err := CreateTags(s.DB, existingTags)
	require.NoError(s.T(), err)

	tags := []string{"example", "foobar"}

	foundTags, err := FindOrCreateTagsByName(s.DB, tags)
	require.NoError(s.T(), err)

	expectedTags := td.Slice(
		[]models.Tag{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Tag{
					Name:   "example",
					Hidden: false,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Tag{
					ID:     returnedTags[0].ID,
					Name:   "foobar",
					Hidden: false,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), foundTags, expectedTags)
}

func (s *TagsSuite) TestFindTagsByName() {
	tags := []models.Tag{
		{
			Name: "nofilter",
		},
		{
			Name: "shotoniphone",
		},
	}

	_, err := CreateTags(s.DB, tags)
	if err != nil {
		s.T().Fatalf("failed to create tags needed for test: %s", err)
	}

	returnedTags, err := FindTagsByName(s.DB, []string{"nofilter"})
	if err != nil {
		s.T().Fatalf("failed get tags: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Tag{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Tag{
					Name: "nofilter",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedTags, expectedResult)
}

func (s *TagsSuite) TestFindTagsByID() {
	tags := []models.Tag{
		{
			Name: "nofilter",
		},
		{
			Name: "shotoniphone",
		},
	}

	persistedTags, err := CreateTags(s.DB, tags)
	require.NoError(s.T(), err)

	returnedTags, err := FindTagsByID(s.DB, []int{persistedTags[0].ID})
	if err != nil {
		s.T().Fatalf("failed get tags: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Tag{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Tag{
					Name: "nofilter",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedTags, expectedResult)
}

func (s *TagsSuite) TestAllTags() {
	tags := []models.Tag{
		{
			Name: "shotoniphone",
		},
		{
			Name: "nofilter",
		},
	}

	_, err := CreateTags(s.DB, tags)
	if err != nil {
		s.T().Fatalf("failed to create tags needed for test: %s", err)
	}

	returnedTags, err := AllTags(s.DB, false, SelectOptions{})
	if err != nil {
		s.T().Fatalf("failed get tags: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Tag{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Tag{
					Name: "shotoniphone",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Tag{
					Name: "nofilter",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedTags, expectedResult)
}

func (s *TagsSuite) TestUpdateTags() {
	initialTags := []models.Tag{
		{
			Name: "shotoniphone",
		},
		{
			Name: "x100f",
		},
	}

	returnedTags, err := CreateTags(s.DB, initialTags)
	if err != nil {
		s.T().Fatalf("failed to create tags needed for test: %s", err)
	}

	expectedTags := td.Slice(
		[]models.Tag{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Tag{
					Name: "shotoniphone",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Tag{
					Name: "x100f",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedTags, expectedTags)

	updatedTags := returnedTags
	updatedTags[0].Name = "iPod"

	returnedTags, err = UpdateTags(s.DB, updatedTags)
	if err != nil {
		s.T().Fatalf("failed to update tags: %s", err)
	}

	expectedTags = td.Slice(
		[]models.Tag{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Tag{
					Name: "ipod",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Tag{
					Name: "x100f",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedTags, expectedTags)

	returnedTags, err = FindTagsByName(s.DB, []string{"ipod"})
	if err != nil {
		s.T().Fatalf("failed get tags: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Tag{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Tag{
					Name: "ipod",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedTags, expectedResult)
}

func (s *TagsSuite) TestDeleteTags() {
	tags := []models.Tag{
		{
			Name: "shotoniphone",
		},
		{
			Name: "x100f",
		},
	}

	returnedTags, err := CreateTags(s.DB, tags)
	if err != nil {
		s.T().Fatalf("failed to create tags: %s", err)
	}

	tagToDelete := returnedTags[0]

	err = DeleteTags(s.DB, []models.Tag{tagToDelete})
	require.NoError(s.T(), err, "unexpected error deleting tags")

	allTags, err := AllTags(s.DB, false, SelectOptions{})
	if err != nil {
		s.T().Fatalf("failed get tags: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Tag{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Tag{
					Name: "x100f",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), allTags, expectedResult)
}
