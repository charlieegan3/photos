package database

import (
	"database/sql"
	"github.com/charlieegan3/photos/cms/internal/pkg/models"
	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"
	"time"
)

type dbPoint struct {
	ID int64 `db:"id"`

	Latitude  float64 `db:"latitude"`
	Longitude float64 `db:"longitude"`
	Altitude  float64 `db:"altitude"`

	Accuracy         float64 `db:"accuracy"`
	VerticalAccuracy float64 `db:"vertical_accuracy"`

	Velocity float64 `db:"velocity"`

	WasOffline bool `db:"was_offline"`

	ImporterID int64 `db:"importer_id"`
	CallerID   int64 `db:"caller_id"`
	ReasonID   int64 `db:"reason_id"`

	ActivityID *int64 `db:"activity_id"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (d *dbPoint) ToRecord(includeID bool) goqu.Record {
	record := goqu.Record{
		"latitude":          d.Latitude,
		"longitude":         d.Longitude,
		"altitude":          d.Altitude,
		"accuracy":          d.Accuracy,
		"vertical_accuracy": d.VerticalAccuracy,
		"velocity":          d.Velocity,
		"was_offline":       d.WasOffline,
		"importer_id":       d.ImporterID,
		"caller_id":         d.CallerID,
		"reason_id":         d.ReasonID,
		"activity_id":       d.ActivityID,
		"created_at":        d.CreatedAt,
		"updated_at":        d.UpdatedAt,
	}

	if includeID {
		record["id"] = d.ID
	}

	return record
}

func newPoint(point dbPoint) models.Point {
	return models.Point{
		ID:               point.ID,
		Latitude:         point.Latitude,
		Longitude:        point.Longitude,
		Altitude:         point.Altitude,
		Accuracy:         point.Accuracy,
		VerticalAccuracy: point.VerticalAccuracy,
		Velocity:         point.Velocity,
		WasOffline:       point.WasOffline,
		ImporterID:       point.ImporterID,
		CallerID:         point.CallerID,
		ReasonID:         point.ReasonID,
		ActivityID:       point.ActivityID,
		CreatedAt:        point.CreatedAt.UTC(),
		UpdatedAt:        point.UpdatedAt,
	}
}

func newDBPoint(point models.Point) dbPoint {
	return dbPoint{
		ID:               point.ID,
		Latitude:         point.Latitude,
		Longitude:        point.Longitude,
		Altitude:         point.Altitude,
		Accuracy:         point.Accuracy,
		VerticalAccuracy: point.VerticalAccuracy,
		Velocity:         point.Velocity,
		WasOffline:       point.WasOffline,
		ImporterID:       point.ImporterID,
		CallerID:         point.CallerID,
		ReasonID:         point.ReasonID,
		ActivityID:       point.ActivityID,
		CreatedAt:        point.CreatedAt,
		UpdatedAt:        point.UpdatedAt,
	}
}

func CreatePoints(
	db *sql.DB,
	importer, caller, reason string,
	activity *models.Activity,
	points []models.Point,
) ([]models.Point, error) {
	var returnedPoints []models.Point

	goquDB := goqu.New("postgres", db)

	tx, err := goquDB.Begin()
	if err != nil {
		return returnedPoints, errors.Wrap(err, "failed to open transaction")
	}

	var importerID, callerID, reasonID int64

	// create importer
	createOrUpdateImporter := tx.Insert("locations.importers").
		Rows(goqu.Record{"name": importer}).
		OnConflict(goqu.DoUpdate("name", goqu.Record{"name": importer})).
		Returning("id").
		Executor()
	_, err = createOrUpdateImporter.ScanVal(&importerID)
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			return returnedPoints, errors.Wrap(rErr, "failed to rollback transaction")
		}
		return returnedPoints, errors.Wrap(err, "failed to find or create importer")
	}

	// create caller
	createOrUpdateCaller := tx.Insert("locations.callers").
		Rows(goqu.Record{"name": caller}).
		OnConflict(goqu.DoUpdate("name", goqu.Record{"name": caller})).
		Returning("id").
		Executor()
	_, err = createOrUpdateCaller.ScanVal(&callerID)
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			return returnedPoints, errors.Wrap(rErr, "failed to rollback transaction")
		}
		return returnedPoints, errors.Wrap(err, "failed to find or create caller")
	}

	// create reason
	createOrUpdateReasons := tx.Insert("locations.reasons").
		Rows(goqu.Record{"reference": reason}).
		OnConflict(goqu.DoUpdate("reference", goqu.Record{"reference": reason})).
		Returning("id").
		Executor()
	_, err = createOrUpdateReasons.ScanVal(&reasonID)
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			return returnedPoints, errors.Wrap(rErr, "failed to rollback transaction")
		}
		return returnedPoints, errors.Wrap(err, "failed to find or create reason")
	}

	// create points
	var pointRecords []goqu.Record
	for _, p := range points {
		dbPoint := newDBPoint(p)
		dbPoint.ReasonID = reasonID
		dbPoint.ImporterID = importerID
		dbPoint.CallerID = callerID
		pointRecords = append(pointRecords, dbPoint.ToRecord(false))
	}

	var dbPoints []dbPoint
	createPoints := tx.Insert("locations.points").
		Rows(pointRecords).
		Returning("*").
		Executor()
	err = createPoints.ScanStructs(&dbPoints)
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			return returnedPoints, errors.Wrap(rErr, "failed to rollback transaction")
		}
		return returnedPoints, errors.Wrap(err, "failed to find or create reason")
	}

	// format points to return
	for _, p := range dbPoints {
		returnedPoints = append(returnedPoints, newPoint(p))
	}

	if err = tx.Commit(); err != nil {
		return returnedPoints, errors.Wrap(err, "failed to commit transaction")
	}

	return returnedPoints, nil
}

func PointsNearTime(db *sql.DB, t time.Time) ([]models.Point, error) {
	var returnedPoints []models.Point

	var dbPoints []dbPoint

	goquDB := goqu.New("postgres", db)

	near := goqu.L("?::timestamp with time zone <-> points.created_at", t.Format(time.RFC3339))

	sub := goquDB.From("locations.points").As("points").
		Select("*", near.As("_")).
		Distinct().
		Order(near.Asc()).
		Limit(5)

	sel := goquDB.From(sub).As("sub").
		Select(
			// TODO, do some magic to not set these manually
			"id",
			"latitude",
			"longitude",
			"altitude",
			"accuracy",
			"vertical_accuracy",
			"velocity",
			"was_offline",
			"importer_id",
			"caller_id",
			"reason_id",
			"activity_id",
			"created_at",
			"updated_at",
		)

	if err := sel.Executor().ScanStructs(&dbPoints); err != nil {
		return returnedPoints, errors.Wrap(err, "failed to select points")
	}

	for _, v := range dbPoints {
		returnedPoints = append(returnedPoints, newPoint(v))
	}

	return returnedPoints, nil
}
