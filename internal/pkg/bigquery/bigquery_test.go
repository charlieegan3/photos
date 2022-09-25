package bigquery

import (
	"context"
	"fmt"
	"google.golang.org/api/option"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/charlieegan3/photos/internal/pkg/models"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/require"
)

func TestPointsInRange(t *testing.T) {
	expected := []models.Point{
		{
			Latitude:  1,
			Longitude: 2,
			CreatedAt: time.Date(2022, 9, 22, 23, 2, 3, 0, time.UTC),
		},
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `
{
  "kind": "result",
  "schema": {
	"fields": [
  {
    "name": "created_at",
    "type": "TIMESTAMP"
  },
  {
    "name": "latitude",
    "type": "FLOAT"
  },
  {
    "name": "longitude",
    "type": "FLOAT"
  },
  {
    "name": "altitude",
    "type": "FLOAT"
  },
  {
    "name": "accuracy",
    "type": "FLOAT"
  },
  {
    "name": "vertical_accuracy",
    "type": "FLOAT"
  },
  {
    "name": "velocity",
    "type": "FLOAT"
  },
  {
    "name": "was_offline",
    "type": "BOOLEAN"
  },
  {
    "name": "importer_id",
    "type": "INTEGER"
  },
  {
    "name": "caller_id",
    "type": "INTEGER"
  },
  {
    "name": "reason_id",
    "type": "INTEGER"
  },
  {
    "mode": "NULLABLE",
    "name": "activity_id",
    "type": "INTEGER"
  }
]
  },
  "jobReference": {
    "projectId": "example",
    "jobId": "example",
    "location":  "example"	
  },
  "totalRows": "1",
  "pageToken": "",
  "rows": [
    {
"f": [
  { "v" : "1663887723" },
  { "v" : "1.0" },
  { "v" : "2.0" },
  { "v" : "0.0" },
  { "v" : "0.0" },
  { "v" : "0.0" },
  { "v" : "0.0" },
  { "v" : "false" },
  { "v" : "0" },
  { "v" : "0" },
  { "v" : "0" },
  { "v" : null }
]
    }
  ],
  "totalBytesProcessed": "100",
  "jobComplete": true,
  "errors": [],
  "cacheHit": false
}`

		w.Write([]byte(response))
	}))
	defer testServer.Close()

	ctx := context.Background()
	client, err := bigquery.NewClient(
		ctx,
		"test-project",
		option.WithEndpoint(testServer.URL),
		option.WithoutAuthentication(),
	)
	require.NoError(t, err)

	actual, err := PointsInRange(
		ctx,
		client,
		"dataset",
		"table",
		time.Date(2022, 1, 1, 1, 2, 3, 0, time.UTC),
		time.Date(2022, 12, 31, 1, 2, 3, 0, time.UTC),
	)
	require.NoError(t, err)

	fmt.Println(actual[0].CreatedAt.Location())

	td.Cmp(t, actual, expected)
}

func TestInsertPoints(t *testing.T) {
	points := []models.Point{
		{
			ID:        1,
			Latitude:  1,
			Longitude: 2,
			CreatedAt: time.Date(2022, 9, 23, 1, 2, 3, 4, time.UTC),
			UpdatedAt: time.Date(2022, 9, 23, 1, 2, 3, 4, time.UTC),
		},
		{
			ID:        2,
			Latitude:  3,
			Longitude: 4,
			CreatedAt: time.Date(2021, 9, 23, 1, 2, 3, 4, time.UTC),
			UpdatedAt: time.Date(2021, 9, 23, 1, 2, 3, 4, time.UTC),
		},
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// response showing that only point 1's time is new
		response := `{
  "kind": "bigquery#tableDataInsertAllResponse",
  "insertErrors": [ ]
}`

		w.Write([]byte(response))
	}))
	defer testServer.Close()

	ctx := context.Background()
	client, err := bigquery.NewClient(
		ctx,
		"test-project",
		option.WithEndpoint(testServer.URL),
		option.WithoutAuthentication(),
	)
	require.NoError(t, err)

	err = InsertPoints(ctx, client, points, "dataset", "table")
	require.NoError(t, err)
}

func TestUnarchivedPoints(t *testing.T) {
	points := []models.Point{
		{
			ID:        1,
			Latitude:  1,
			Longitude: 2,
			CreatedAt: time.Date(2022, 9, 23, 1, 2, 3, 4, time.UTC),
			UpdatedAt: time.Date(2022, 9, 23, 1, 2, 3, 4, time.UTC),
		},
		{
			ID:        2,
			Latitude:  3,
			Longitude: 4,
			CreatedAt: time.Date(2021, 9, 23, 1, 2, 3, 4, time.UTC),
			UpdatedAt: time.Date(2021, 9, 23, 1, 2, 3, 4, time.UTC),
		},
	}

	expected := []models.Point{
		{
			ID:        1,
			Latitude:  1,
			Longitude: 2,
			CreatedAt: time.Date(2022, 9, 23, 1, 2, 3, 4, time.UTC),
			UpdatedAt: time.Date(2022, 9, 23, 1, 2, 3, 4, time.UTC),
		},
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// response showing that only point 1's time is new
		response := `
{
  "kind": "result",
  "schema": {
	"fields": [
  {
    "name": "created_at",
    "type": "TIMESTAMP"
  }
]

  },
  "jobReference": {
    "projectId": "example",
    "jobId": "example",
    "location":  "example"	
  },
  "totalRows": "1",
  "pageToken": "",
  "rows": [
    {
"f": [
{
	  "v" : "1663894923"
}
]
    }
  ],
  "totalBytesProcessed": "100",
  "jobComplete": true,
  "errors": [],
  "cacheHit": false
}`

		w.Write([]byte(response))
	}))
	defer testServer.Close()

	ctx := context.Background()
	client, err := bigquery.NewClient(
		ctx,
		"test-project",
		option.WithEndpoint(testServer.URL),
		option.WithoutAuthentication(),
	)
	require.NoError(t, err)

	actual, err := UnarchivedPoints(ctx, client, points, "dataset", "table")
	require.NoError(t, err)

	td.Cmp(t, actual, expected)
}
