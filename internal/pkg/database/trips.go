package database

import (
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

func newTrip(trip dbTrip) models.Trip {
	return models.Trip{
		ID: trip.ID,

		Title:       trip.Title,
		Description: trip.Description,

		StartDate: trip.StartDate.UTC(),
		EndDate:   trip.EndDate.UTC(),

		CreatedAt: trip.CreatedAt,
		UpdatedAt: trip.UpdatedAt,
	}
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

func CreateTrips(db *sql.DB, trips []models.Trip) (results []models.Trip, err error) {
	records := []goqu.Record{}
	for _, v := range trips {
		d := newDBTrip(v)
		records = append(records, d.ToRecord(false))
	}

	var dbTrips []dbTrip

	goquDB := goqu.New("postgres", db)
	insert := goquDB.Insert("trips").Returning(goqu.Star()).Rows(records).Executor()
	if err := insert.ScanStructs(&dbTrips); err != nil {
		return results, errors.Wrap(err, "failed to insert trips")
	}

	for _, v := range dbTrips {
		results = append(results, newTrip(v))
	}

	return results, nil
}

func FindTripsByID(db *sql.DB, id []int) (results []models.Trip, err error) {
	var dbTrips []dbTrip

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("trips").Select("*").Where(goqu.Ex{"id": id}).Executor()
	if err := insert.ScanStructs(&dbTrips); err != nil {
		return results, errors.Wrap(err, "failed to select trips by id")
	}

	for _, v := range dbTrips {
		results = append(results, newTrip(v))
	}

	return results, nil
}

func AllTrips(db *sql.DB) (results []models.Trip, err error) {
	var dbTrips []dbTrip

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("trips").Select("*").Order(goqu.I("start_date").Desc()).Executor()

	if err := insert.ScanStructs(&dbTrips); err != nil {
		return results, errors.Wrap(err, "failed to select trips")
	}

	// this is needed in case there are no items added, we don't want to return
	// nil but rather an empty slice
	results = []models.Trip{}
	for _, v := range dbTrips {
		results = append(results, newTrip(v))
	}

	return results, nil
}

func DeleteTrips(db *sql.DB, trips []models.Trip) (err error) {
	var ids []int
	for _, d := range trips {
		ids = append(ids, d.ID)
	}

	goquDB := goqu.New("postgres", db)
	del, _, err := goquDB.Delete("trips").Where(
		goqu.Ex{"id": ids},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("failed to build trips delete query: %s", err)
	}
	_, err = db.Exec(del)
	if err != nil {
		return fmt.Errorf("failed to delete trips: %s", err)
	}

	return nil
}

// UpdateTrips is not implemented as a single SQL query since update many in
// place is not supported by goqu and it wasn't worth the work (TODO)
func UpdateTrips(db *sql.DB, trips []models.Trip) (results []models.Trip, err error) {
	records := []goqu.Record{}
	for _, v := range trips {
		d := newDBTrip(v)
		records = append(records, d.ToRecord(true))
	}

	goquDB := goqu.New("postgres", db)
	tx, err := goquDB.Begin()
	if err != nil {
		return results, errors.Wrap(err, "failed to open tx for updating trips")
	}

	for _, record := range records {
		var result dbTrip
		update := tx.From("trips").
			Where(goqu.Ex{"id": record["id"]}).
			Update().
			Set(record).
			Returning(goqu.Star()).
			Executor()
		if _, err = update.ScanStruct(&result); err != nil {
			if rErr := tx.Rollback(); rErr != nil {
				return results, errors.Wrap(err, "failed to rollback")
			}
			return results, errors.Wrap(err, "failed to update, rolled back")
		}

		results = append(results, newTrip(result))
	}
	if err = tx.Commit(); err != nil {
		return results, errors.Wrap(err, "failed to commit transaction")
	}

	return results, nil
}
