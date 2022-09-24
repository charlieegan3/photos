package bigquery

import (
	"fmt"
	"golang.org/x/net/context"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/charlieegan3/photos/internal/pkg/models"
	"google.golang.org/api/iterator"
)

type bqPoint struct {
	Latitude  float64 `bigquery:"latitude"`
	Longitude float64 `bigquery:"longitude"`
	Altitude  float64 `bigquery:"altitude"`

	Accuracy         float64 `bigquery:"accuracy"`
	VerticalAccuracy float64 `bigquery:"vertical_accuracy"`

	Velocity float64 `bigquery:"velocity"`

	WasOffline bool `bigquery:"was_offline"`

	ImporterID int64 `bigquery:"importer_id"`
	CallerID   int64 `bigquery:"caller_id"`
	ReasonID   int64 `bigquery:"reason_id"`

	ActivityID bigquery.NullInt64 `bigquery:"activity_id"`

	CreatedAt time.Time `bigquery:"created_at"`
}

func newBqPoint(point models.Point) bqPoint {

	// activityID is the only nullable field
	activityID := bigquery.NullInt64{
		Int64: 0,
		Valid: false,
	}

	if point.ActivityID != nil {
		activityID.Valid = true
		activityID.Int64 = *point.ActivityID
	}

	return bqPoint{
		Latitude:  point.Latitude,
		Longitude: point.Longitude,
		Altitude:  point.Altitude,

		Accuracy:         point.Accuracy,
		VerticalAccuracy: point.VerticalAccuracy,

		Velocity: point.Velocity,

		WasOffline: point.WasOffline,

		ImporterID: point.ImporterID,
		CallerID:   point.CallerID,
		ReasonID:   point.ReasonID,

		ActivityID: activityID,

		CreatedAt: point.CreatedAt,
	}
}

func InsertPoints(
	ctx context.Context,
	client *bigquery.Client,
	points []models.Point,
	dataset, table string,
) error {
	inserter := client.Dataset(dataset).Table(table).Inserter()

	var bqPoints []bqPoint
	for _, p := range points {
		bqPoints = append(bqPoints, newBqPoint(p))
	}

	return inserter.Put(ctx, bqPoints)
}

// UnarchivedPoints takes a list of points with created_at timestamps. If the timestamp of a point is not found in the
// archive, then it is returned. This is so that new points can be added to the archive.
func UnarchivedPoints(
	ctx context.Context,
	client *bigquery.Client,
	points []models.Point,
	dataset, table string,
) ([]models.Point, error) {

	var unarchivedPoints []models.Point

	// create a list of timestamps in the SQL format to use in the query
	var timestamps []string
	for _, p := range points {
		timestamps = append(
			timestamps,
			fmt.Sprintf(
				"TIMESTAMP_SECONDS('%d')",
				p.CreatedAt.Unix(),
			),
		)
	}

	queryString := fmt.Sprintf(
		`WITH
  new_data AS (
  SELECT
    *
  FROM
    UNNEST([%s]) AS created_at)
SELECT
  *
FROM
  new_data
WHERE
  created_at NOT IN (
  SELECT
    created_at
  FROM
    %s.%s)`,
		strings.Join(timestamps, ","),
		dataset,
		table,
	)
	q := client.Query(queryString)
	it, err := q.Read(ctx)
	if err != nil {
		return unarchivedPoints, fmt.Errorf("failed query for new timestamps: %w", err)
	}

	var newTimestamps []time.Time
	for {
		var values []bigquery.Value
		err := it.Next(&values)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return unarchivedPoints, fmt.Errorf("failed reading results: %w", err)
		}

		if len(values) != 1 {
			return unarchivedPoints, fmt.Errorf("unexpected number of values in row: %w", err)
		}

		t, ok := values[0].(time.Time)
		if !ok {
			return unarchivedPoints, fmt.Errorf("unexpected type for time value in row: %w", err)
		}

		newTimestamps = append(newTimestamps, t)
	}

	// if points have a matching time in the new timestamps, then we need to return them as Unarchived
	for _, p := range points {
		for _, t := range newTimestamps {
			if p.CreatedAt.Unix() == t.Unix() {
				unarchivedPoints = append(unarchivedPoints, p)
				break
			}
		}
	}

	return unarchivedPoints, nil
}
