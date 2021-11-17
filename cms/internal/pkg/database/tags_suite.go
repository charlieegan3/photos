package database

import (
	"database/sql"

	"github.com/charlieegan3/photos/cms/internal/pkg/models"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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
			Name: "No Filter",
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
					Name: "no-filter",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Tag{
					Name: "shotoniphone",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedTags, expectedTags)
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

	returnedTags, err := FindTagsByName(s.DB, "nofilter")
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

	returnedTags, err := AllTags(s.DB)
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

	returnedTags, err = FindTagsByName(s.DB, "ipod")
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

	allTags, err := AllTags(s.DB)
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
