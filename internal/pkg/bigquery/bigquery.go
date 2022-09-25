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

func newPointFromBqValues(values []bigquery.Value) (models.Point, error) {
	createdAt, ok := values[0].(time.Time)
	if !ok {
		return models.Point{}, fmt.Errorf("failed to parse created_at")
	}
	latitude, ok := values[1].(float64)
	if !ok {
		return models.Point{}, fmt.Errorf("failed to parse latitude")
	}
	longitude, ok := values[2].(float64)
	if !ok {
		return models.Point{}, fmt.Errorf("failed to parse longitude")
	}
	altitude, ok := values[3].(float64)
	if !ok {
		return models.Point{}, fmt.Errorf("failed to parse altitude")
	}
	accuracy, ok := values[4].(float64)
	if !ok {
		return models.Point{}, fmt.Errorf("failed to parse accuracy")
	}
	verticalAccuracy, ok := values[5].(float64)
	if !ok {
		return models.Point{}, fmt.Errorf("failed to parse vertical_accuracy")
	}
	velocity, ok := values[6].(float64)
	if !ok {
		return models.Point{}, fmt.Errorf("failed to parse velocity")
	}
	offline, ok := values[7].(bool)
	if !ok {
		return models.Point{}, fmt.Errorf("failed to parse was_offline")
	}
	importer, ok := values[8].(int64)
	if !ok {
		return models.Point{}, fmt.Errorf("failed to parse importer_id")
	}
	caller, ok := values[9].(int64)
	if !ok {
		return models.Point{}, fmt.Errorf("failed to parse caller_id")
	}
	reason, ok := values[10].(int64)
	if !ok {
		return models.Point{}, fmt.Errorf("failed to parse reason_id")
	}

	var activity *int64
	if values[11] != nil {
		activityNullable, ok := values[11].(bigquery.NullInt64)
		if !ok {
			return models.Point{}, fmt.Errorf("failed to parse activity_id")
		}
		if activityNullable.Valid {
			activity = &activityNullable.Int64
		}
	}

	return models.Point{
		ID:               0,
		Latitude:         latitude,
		Longitude:        longitude,
		Altitude:         altitude,
		Accuracy:         accuracy,
		VerticalAccuracy: verticalAccuracy,
		Velocity:         velocity,
		WasOffline:       offline,
		ImporterID:       importer,
		CallerID:         caller,
		ReasonID:         reason,
		ActivityID:       activity,
		CreatedAt:        createdAt,
	}, nil
}

func PointsInRange(
	ctx context.Context,
	client *bigquery.Client,
	dataset string,
	table string,
	notBefore, notAfter time.Time,
) ([]models.Point, error) {

	var points []models.Point

	queryString := fmt.Sprintf(
		`SELECT
  *
FROM
  %s.%s
WHERE
  created_at BETWEEN TIMESTAMP_SECONDS(%d)
  AND TIMESTAMP_SECONDS(%d)
ORDER BY
  created_at ASC`,
		dataset,
		table,
		notBefore.Unix(), notAfter.Unix(),
	)
	q := client.Query(queryString)
	it, err := q.Read(ctx)
	if err != nil {
		return points, fmt.Errorf("failed query for points: %w", err)
	}

	for {
		var values []bigquery.Value
		err := it.Next(&values)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return points, fmt.Errorf("failed reading results: %w", err)
		}

		point, err := newPointFromBqValues(values)
		if err != nil {
			return points, fmt.Errorf("failed parsing point: %w", err)
		}

		points = append(points, point)
	}

	return points, nil
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
				"TIMESTAMP_SECONDS(%d)",
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
