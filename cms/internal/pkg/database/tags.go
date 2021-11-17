package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/charlieegan3/photos/cms/internal/pkg/models"
	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"
)

type dbTag struct {
	ID     int    `db:"id"`
	Name   string `db:"name"`
	Hidden bool   `db:"hidden"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (d *dbTag) ToRecord(includeID bool) goqu.Record {
	record := goqu.Record{
		"name":   d.Name,
		"hidden": d.Hidden,
	}

	if includeID {
		record["id"] = d.ID
	}

	return record
}

func newTag(tag dbTag) models.Tag {
	return models.Tag{
		ID:        tag.ID,
		Name:      tag.Name,
		Hidden:    tag.Hidden,
		CreatedAt: tag.CreatedAt,
		UpdatedAt: tag.UpdatedAt,
	}
}

func newDBTag(tag models.Tag) dbTag {
	return dbTag{
		ID:        tag.ID,
		Name:      tag.Name,
		Hidden:    tag.Hidden,
		CreatedAt: tag.CreatedAt,
		UpdatedAt: tag.UpdatedAt,
	}
}

func CreateTags(db *sql.DB, tags []models.Tag) (results []models.Tag, err error) {
	records := []goqu.Record{}
	for _, v := range tags {
		d := newDBTag(v)
		records = append(records, d.ToRecord(false))
	}

	var dbTags []dbTag

	goquDB := goqu.New("postgres", db)
	insert := goquDB.Insert("tags").Returning(goqu.Star()).Rows(records).Executor()
	if err := insert.ScanStructs(&dbTags); err != nil {
		return results, errors.Wrap(err, "failed to insert tags")
	}

	for _, v := range dbTags {
		results = append(results, newTag(v))
	}

	return results, nil
}

func FindTagsByName(db *sql.DB, name string) (results []models.Tag, err error) {
	var dbTags []dbTag

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("tags").Select("*").Where(goqu.Ex{"name": name}).Executor()
	if err := insert.ScanStructs(&dbTags); err != nil {
		return results, errors.Wrap(err, "failed to select tags by slug")
	}

	for _, v := range dbTags {
		results = append(results, newTag(v))
	}

	return results, nil
}

func AllTags(db *sql.DB) (results []models.Tag, err error) {
	var dbTags []dbTag

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("tags").Select("*").Executor()
	if err := insert.ScanStructs(&dbTags); err != nil {
		return results, errors.Wrap(err, "failed to select tags")
	}

	// this is needed in case there are no items added, we don't want to return
	// nil but rather an empty slice
	results = []models.Tag{}
	for _, v := range dbTags {
		results = append(results, newTag(v))
	}

	return results, nil
}

func DeleteTags(db *sql.DB, tags []models.Tag) (err error) {
	var ids []int
	for _, d := range tags {
		ids = append(ids, d.ID)
	}

	goquDB := goqu.New("postgres", db)
	del, _, err := goquDB.Delete("tags").Where(
		goqu.Ex{"id": ids},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("failed to build tags delete query: %s", err)
	}
	_, err = db.Exec(del)
	if err != nil {
		return fmt.Errorf("failed to delete tags: %s", err)
	}

	return nil
}

// UpdateTags is not implemented as a single SQL query since update many in
// place is not supported by goqu and it wasn't worth the work (TODO)
func UpdateTags(db *sql.DB, tags []models.Tag) (results []models.Tag, err error) {
	records := []goqu.Record{}
	for _, v := range tags {
		d := newDBTag(v)
		records = append(records, d.ToRecord(true))
	}

	goquDB := goqu.New("postgres", db)
	tx, err := goquDB.Begin()
	if err != nil {
		return results, errors.Wrap(err, "failed to open tx for updating tags")
	}

	for _, record := range records {
		var result dbTag
		update := tx.From("tags").
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

		results = append(results, newTag(result))
	}
	if err = tx.Commit(); err != nil {
		return results, errors.Wrap(err, "failed to commit transaction")
	}

	return results, nil
}
