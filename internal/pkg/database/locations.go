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

func FindLocationsByID(db *sql.DB, id []int) (results []models.Location, err error) {
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

func FindLocationsByName(db *sql.DB, name string) (results []models.Location, err error) {
	var dbLocations []dbLocation

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("locations").Select("*").Where(goqu.Ex{"name": name}).Executor()
	if err := insert.ScanStructs(&dbLocations); err != nil {
		return results, errors.Wrap(err, "failed to select locations by name")
	}

	for _, v := range dbLocations {
		results = append(results, newLocation(v))
	}

	return results, nil
}

func AllLocations(db *sql.DB) (results []models.Location, err error) {
	var dbLocations []dbLocation

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("locations").Select("*").Order(goqu.I("name").Asc()).Executor()

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

func MergeLocations(db *sql.DB, locationName, oldLocationName string) (id int, err error) {
	goquDB := goqu.New("postgres", db)

	var oldID int

	tx, err := goquDB.Begin()
	if err != nil {
		return 0, errors.Wrap(err, "failed to open transaction")
	}

	newLocationID := tx.From("locations").
		Select("id").
		Where(goqu.Ex{"name": locationName}).Executor()
	result, err := newLocationID.ScanVal(&id)
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			return 0, errors.Wrap(rErr, "failed to rollback transaction")
		}
		return 0, errors.Wrap(err, "failed to get new location ID")
	}
	if !result {
		if rErr := tx.Rollback(); rErr != nil {
			return 0, errors.Wrap(rErr, "failed to rollback transaction")
		}
		return 0, errors.Wrap(err, "failed to get new location from queried row")
	}

	oldLocationID := tx.From("locations").
		Select("id").
		Where(goqu.Ex{"name": oldLocationName}).Executor()
	result, err = oldLocationID.ScanVal(&oldID)
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			return 0, errors.Wrap(rErr, "failed to rollback transaction")
		}
		return 0, errors.Wrap(err, "failed to get old location ID")
	}
	if !result {
		if rErr := tx.Rollback(); rErr != nil {
			return 0, errors.Wrap(rErr, "failed to rollback transaction")
		}
		return 0, errors.Wrap(err, "failed to get old location from queried row")
	}

	updatePosts := tx.Update("posts").
		Where(goqu.Ex{"location_id": oldID}).
		Set(map[string]interface{}{"location_id": id}).
		Executor()
	_, err = updatePosts.Exec()
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			return 0, errors.Wrap(rErr, "failed to rollback transaction")
		}
		return 0, errors.Wrap(err, "failed to update posts to merged location")
	}

	deleteLocation := tx.Delete("locations").
		Where(goqu.Ex{"id": oldID}).Executor()
	_, err = deleteLocation.Exec()
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			return 0, errors.Wrap(rErr, "failed to rollback transaction")
		}
		return 0, errors.Wrap(err, "failed to delete old location")
	}

	if err = tx.Commit(); err != nil {
		return 0, errors.Wrap(err, "failed to commit transaction")
	}

	return id, nil
}

func NearbyLocations(db *sql.DB, lat, lon float64) (results []models.Location, err error) {
	var dbLocations []dbLocation

	goquDB := goqu.New("postgres", db)
	sub1 := goquDB.From(goqu.T("locations")).
		Select("*", goqu.L("calculate_distance(?,?, locations.latitude,locations.longitude, 'K')", lat, lon))

	sub2 := goquDB.From(sub1).
		Select("*").
		Where(goqu.I("calculate_distance").Lt(10)).
		Order(goqu.I("calculate_distance").Asc()).
		Limit(10)

	sel := goquDB.From(sub2).
		Select("id",
			"name",
			"slug",
			"latitude",
			"longitude",
			"created_at",
			"updated_at").
		Limit(10)

	if err := sel.Executor().ScanStructs(&dbLocations); err != nil {
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
