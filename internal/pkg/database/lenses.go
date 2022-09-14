package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"

	"github.com/charlieegan3/photos/internal/pkg/models"
)

type dbLens struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (d *dbLens) ToRecord(includeID bool) goqu.Record {
	record := goqu.Record{
		"name": d.Name,
	}

	if includeID {
		record["id"] = d.ID
	}

	return record
}

func newLens(lens dbLens) models.Lens {
	return models.Lens{
		ID:        lens.ID,
		Name:      lens.Name,
		CreatedAt: lens.CreatedAt,
		UpdatedAt: lens.UpdatedAt,
	}
}

func newDBLens(lens models.Lens) dbLens {
	return dbLens{
		ID:        lens.ID,
		Name:      lens.Name,
		CreatedAt: lens.CreatedAt,
		UpdatedAt: lens.UpdatedAt,
	}
}

func CreateLenses(db *sql.DB, lenses []models.Lens) (results []models.Lens, err error) {
	records := []goqu.Record{}
	for _, v := range lenses {
		d := newDBLens(v)
		records = append(records, d.ToRecord(false))
	}

	var dbLenses []dbLens

	goquDB := goqu.New("postgres", db)
	insert := goquDB.Insert("lenses").Returning(goqu.Star()).Rows(records).Executor()
	if err := insert.ScanStructs(&dbLenses); err != nil {
		return results, errors.Wrap(err, "failed to insert lenses")
	}

	for _, v := range dbLenses {
		results = append(results, newLens(v))
	}

	return results, nil
}

func FindLensesByID(db *sql.DB, id []int64) (results []models.Lens, err error) {
	var dbLenses []dbLens

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("lenses").Select("*").Where(goqu.Ex{"id": id}).Executor()
	if err := insert.ScanStructs(&dbLenses); err != nil {
		return results, errors.Wrap(err, "failed to select lenses by slug")
	}

	for _, v := range dbLenses {
		results = append(results, newLens(v))
	}

	return results, nil
}

func FindLensesByName(db *sql.DB, name string) (results []models.Lens, err error) {
	var dbLenses []dbLens

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("lenses").Select("*").Where(goqu.Ex{"name": name}).Executor()
	if err := insert.ScanStructs(&dbLenses); err != nil {
		return results, errors.Wrap(err, "failed to select lenses by slug")
	}

	for _, v := range dbLenses {
		results = append(results, newLens(v))
	}

	return results, nil
}

func AllLenses(db *sql.DB) (results []models.Lens, err error) {
	var dbLenses []dbLens

	goquDB := goqu.New("postgres", db)
	selectLenses := goquDB.From("lenses").
		Select("*").
		Order(goqu.I("name").Asc()).
		Executor()
	if err := selectLenses.ScanStructs(&dbLenses); err != nil {
		return results, errors.Wrap(err, "failed to select lenses")
	}

	// this is needed in case there are no items added, we don't want to return
	// nil but rather an empty slice
	results = []models.Lens{}
	for _, v := range dbLenses {
		results = append(results, newLens(v))
	}

	return results, nil
}

func MostRecentlyUsedLens(db *sql.DB) (result models.Lens, err error) {
	var dbLenses []dbLens

	goquDB := goqu.New("postgres", db)
	selectLenses := goquDB.From("lenses").
		InnerJoin(goqu.T("medias"), goqu.On(goqu.Ex{"medias.lens_id": goqu.I("lenses.id")})).
		Select("lenses.*").
		Order(goqu.I("medias.taken_at").Desc()).
		Executor()
	if err := selectLenses.ScanStructs(&dbLenses); err != nil {
		return result, errors.Wrap(err, "failed to select lenses")
	}

	// this is needed in case there are no items added, we don't want to return
	// nil but rather an empty slice
	results := []models.Lens{}
	for _, v := range dbLenses {
		results = append(results, newLens(v))
	}

	if len(results) < 1 {
		return result, nil
	}

	result = results[0]
	return result, nil
}

func DeleteLenses(db *sql.DB, lenses []models.Lens) (err error) {
	var ids []int64
	for _, d := range lenses {
		ids = append(ids, d.ID)
	}

	goquDB := goqu.New("postgres", db)
	del, _, err := goquDB.Delete("lenses").Where(
		goqu.Ex{"id": ids},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("failed to build lenses delete query: %s", err)
	}
	_, err = db.Exec(del)
	if err != nil {
		return fmt.Errorf("failed to delete lenses: %s", err)
	}

	return nil
}

// UpdateLenses is not implemented as a single SQL query since update many in
// place is not supported by goqu and it wasn't worth the work (TODO)
func UpdateLenses(db *sql.DB, lenses []models.Lens) (results []models.Lens, err error) {
	records := []goqu.Record{}
	for _, v := range lenses {
		d := newDBLens(v)
		records = append(records, d.ToRecord(true))
	}

	goquDB := goqu.New("postgres", db)
	tx, err := goquDB.Begin()
	if err != nil {
		return results, errors.Wrap(err, "failed to open tx for updating lenses")
	}

	for _, record := range records {
		var result dbLens
		update := tx.From("lenses").
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

		results = append(results, newLens(result))
	}
	if err = tx.Commit(); err != nil {
		return results, errors.Wrap(err, "failed to commit transaction")
	}

	return results, nil
}
