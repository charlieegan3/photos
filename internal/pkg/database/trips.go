package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"

	"github.com/charlieegan3/photos/internal/pkg/models"
)

type dbTrip struct {
	ID int `db:"id"`

	Title       string `db:"title"`
	Description string `db:"description"`

	StartDate time.Time `db:"start_date"`
	EndDate   time.Time `db:"end_date"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (d *dbTrip) ToRecord(includeID bool) goqu.Record {
	record := goqu.Record{
		"title":       d.Title,
		"description": d.Description,
		"start_date":  d.StartDate.UTC(),
		"end_date":    d.EndDate.UTC(),
	}

	if includeID {
		record["id"] = d.ID
	}

	return record
}

func (d *dbTrip) ToModel() models.Trip {
	return models.Trip{
		ID: d.ID,

		Title:       d.Title,
		Description: d.Description,

		StartDate: d.StartDate.UTC(),
		EndDate:   d.EndDate.UTC(),

		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func newTrip(trip dbTrip) models.Trip {
	return (&trip).ToModel()
}

func newDBTrip(trip models.Trip) dbTrip {
	return dbTrip{
		ID: trip.ID,

		Title:       trip.Title,
		Description: trip.Description,

		StartDate: trip.StartDate.UTC(),
		EndDate:   trip.EndDate.UTC(),

		CreatedAt: trip.CreatedAt,
		UpdatedAt: trip.UpdatedAt,
	}
}

func CreateTrips(ctx context.Context, db *sql.DB, trips []models.Trip) (results []models.Trip, err error) {
	records := []goqu.Record{}
	for i := range trips {
		d := newDBTrip(trips[i])
		records = append(records, d.ToRecord(false))
	}

	var dbTrips []dbTrip

	goquDB := goqu.New("postgres", db)
	insert := goquDB.Insert("photos.trips").Returning(goqu.Star()).Rows(records).Executor()
	err = insert.ScanStructsContext(ctx, &dbTrips)
	if err != nil {
		return results, errors.Wrap(err, "failed to insert trips")
	}

	for i := range dbTrips {
		results = append(results, newTrip(dbTrips[i]))
	}

	return results, nil
}

func FindTripsByID(ctx context.Context, db *sql.DB, id []int) (results []models.Trip, err error) {
	var dbTrips []dbTrip

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("photos.trips").Select("*").Where(goqu.Ex{"id": id}).Executor()
	err = insert.ScanStructsContext(ctx, &dbTrips)
	if err != nil {
		return results, errors.Wrap(err, "failed to select trips by id")
	}

	for i := range dbTrips {
		results = append(results, newTrip(dbTrips[i]))
	}

	return results, nil
}

func AllTrips(ctx context.Context, db *sql.DB) (results []models.Trip, err error) {
	var dbTrips []dbTrip

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("photos.trips").Select("*").Order(goqu.I("start_date").Desc()).Executor()

	err = insert.ScanStructsContext(ctx, &dbTrips)
	if err != nil {
		return results, errors.Wrap(err, "failed to select trips")
	}

	// this is needed in case there are no items added, we don't want to return
	// nil but rather an empty slice
	results = []models.Trip{}
	for i := range dbTrips {
		results = append(results, newTrip(dbTrips[i]))
	}

	return results, nil
}

func DeleteTrips(ctx context.Context, db *sql.DB, trips []models.Trip) (err error) {
	ids := make([]int, len(trips))
	for i := range trips {
		ids[i] = trips[i].ID
	}

	goquDB := goqu.New("postgres", db)
	del, _, err := goquDB.Delete("photos.trips").Where(
		goqu.Ex{"id": ids},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("failed to build trips delete query: %w", err)
	}
	_, err = db.ExecContext(ctx, del)
	if err != nil {
		return fmt.Errorf("failed to delete trips: %w", err)
	}

	return nil
}

// UpdateTrips is not implemented as a single SQL query since update many in
// place is not supported by goqu and it wasn't worth the work (TODO).
func UpdateTrips(ctx context.Context, db *sql.DB, trips []models.Trip) ([]models.Trip, error) {
	return BulkUpdate(ctx, db, "photos.trips", trips, newDBTrip)
}
