package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"

	"github.com/charlieegan3/photos/cms/internal/pkg/models"
)

type dbLocation struct {
	ID        int     `db:"id"`
	Name      string  `db:"name"`
	Slug      string  `db:"slug"`
	Latitude  float64 `db:"latitude"`
	Longitude float64 `db:"longitude"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (d *dbLocation) ToRecord(includeID bool) goqu.Record {
	record := goqu.Record{
		"name":      d.Name,
		"latitude":  d.Latitude,
		"longitude": d.Longitude,
	}

	if includeID {
		record["id"] = d.ID
	}

	return record
}

func newLocation(location dbLocation) models.Location {
	return models.Location{
		ID:        location.ID,
		Name:      location.Name,
		Slug:      location.Slug,
		Latitude:  location.Latitude,
		Longitude: location.Longitude,
		CreatedAt: location.CreatedAt,
		UpdatedAt: location.UpdatedAt,
	}
}

func newDBLocation(location models.Location) dbLocation {
	return dbLocation{
		ID:        location.ID,
		Name:      location.Name,
		Slug:      location.Slug,
		Latitude:  location.Latitude,
		Longitude: location.Longitude,
		CreatedAt: location.CreatedAt,
		UpdatedAt: location.UpdatedAt,
	}
}

func CreateLocations(db *sql.DB, locations []models.Location) (results []models.Location, err error) {
	records := []goqu.Record{}
	for _, v := range locations {
		d := newDBLocation(v)
		records = append(records, d.ToRecord(false))
	}

	var dbLocations []dbLocation

	goquDB := goqu.New("postgres", db)
	insert := goquDB.Insert("locations").Returning(goqu.Star()).Rows(records).Executor()
	if err := insert.ScanStructs(&dbLocations); err != nil {
		return results, errors.Wrap(err, "failed to insert locations")
	}

	for _, v := range dbLocations {
		results = append(results, newLocation(v))
	}

	return results, nil
}

func FindLocationsByID(db *sql.DB, id int) (results []models.Location, err error) {
	var dbLocations []dbLocation

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("locations").Select("*").Where(goqu.Ex{"id": id}).Executor()
	if err := insert.ScanStructs(&dbLocations); err != nil {
		return results, errors.Wrap(err, "failed to select locations by id")
	}

	for _, v := range dbLocations {
		results = append(results, newLocation(v))
	}

	return results, nil
}

func AllLocations(db *sql.DB) (results []models.Location, err error) {
	var dbLocations []dbLocation

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("locations").Select("*").Executor()
	if err := insert.ScanStructs(&dbLocations); err != nil {
		return results, errors.Wrap(err, "failed to select locations")
	}

	// this is needed in case there are no items added, we don't want to return
	// nil but rather an empty slice
	results = []models.Location{}
	for _, v := range dbLocations {
		results = append(results, newLocation(v))
	}

	return results, nil
}

func DeleteLocations(db *sql.DB, locations []models.Location) (err error) {
	var ids []int
	for _, d := range locations {
		ids = append(ids, d.ID)
	}

	goquDB := goqu.New("postgres", db)
	del, _, err := goquDB.Delete("locations").Where(
		goqu.Ex{"id": ids},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("failed to build locations delete query: %s", err)
	}
	_, err = db.Exec(del)
	if err != nil {
		return fmt.Errorf("failed to delete locations: %s", err)
	}

	return nil
}

// UpdateLocations is not implemented as a single SQL query since update many in
// place is not supported by goqu and it wasn't worth the work (TODO)
func UpdateLocations(db *sql.DB, locations []models.Location) (results []models.Location, err error) {
	records := []goqu.Record{}
	for _, v := range locations {
		d := newDBLocation(v)
		records = append(records, d.ToRecord(true))
	}

	goquDB := goqu.New("postgres", db)
	tx, err := goquDB.Begin()
	if err != nil {
		return results, errors.Wrap(err, "failed to open tx for updating locations")
	}

	for _, record := range records {
		var result dbLocation
		update := tx.From("locations").
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

		results = append(results, newLocation(result))
	}
	if err = tx.Commit(); err != nil {
		return results, errors.Wrap(err, "failed to commit transaction")
	}

	return results, nil
}
