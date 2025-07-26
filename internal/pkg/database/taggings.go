package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/pkg/errors"

	"github.com/charlieegan3/photos/internal/pkg/models"
)

type dbTagging struct {
	ID int `db:"id"`

	PostID int `db:"post_id"`
	TagID  int `db:"tag_id"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (d *dbTagging) ToRecord(includeID bool) goqu.Record {
	record := goqu.Record{
		"post_id": d.PostID,
		"tag_id":  d.TagID,
	}

	if includeID {
		record["id"] = d.ID
	}

	return record
}

func newTagging(tagging dbTagging) models.Tagging {
	return models.Tagging{
		ID: tagging.ID,

		CreatedAt: tagging.CreatedAt,
		UpdatedAt: tagging.UpdatedAt,

		PostID: tagging.PostID,
		TagID:  tagging.TagID,
	}
}

func newDBTagging(tagging models.Tagging) dbTagging {
	return dbTagging{
		ID: tagging.ID,

		CreatedAt: tagging.CreatedAt,
		UpdatedAt: tagging.UpdatedAt,

		PostID: tagging.PostID,
		TagID:  tagging.TagID,
	}
}

func CreateTaggings(ctx context.Context, db *sql.DB, taggings []models.Tagging) (results []models.Tagging, err error) {
	records := []goqu.Record{}
	for _, v := range taggings {
		d := newDBTagging(v)
		records = append(records, d.ToRecord(false))
	}

	var dbTaggings []dbTagging

	goquDB := goqu.New("postgres", db)
	insert := goquDB.Insert("photos.taggings").
		Returning(goqu.Star()).
		Rows(records).
		OnConflict(goqu.DoNothing()). // there are only two fields
		Executor()
	if err := insert.ScanStructsContext(ctx, &dbTaggings); err != nil {
		return results, errors.Wrap(err, "failed to insert taggings")
	}

	for _, v := range dbTaggings {
		results = append(results, newTagging(v))
	}

	return results, nil
}

func FindOrCreateTaggings(
	ctx context.Context,
	db *sql.DB,
	taggings []models.Tagging,
) (results []models.Tagging, err error) {
	var ex []exp.Expression
	for _, t := range taggings {
		ex = append(ex, goqu.Ex{
			"post_id": t.PostID,
			"tag_id":  t.TagID,
		})
	}

	var dbTaggings []dbTagging

	goquDB := goqu.New("postgres", db)
	sel := goquDB.From("photos.taggings").Select("*").Where(goqu.Or(ex...)).Executor()
	if err := sel.ScanStructsContext(ctx, &dbTaggings); err != nil {
		return results, errors.Wrap(err, "failed to select taggings by post_id")
	}

	for _, v := range dbTaggings {
		results = append(results, newTagging(v))
	}

	var taggingsToCreate []models.Tagging
	for _, t := range taggings {
		found := false
		for _, tagging := range dbTaggings {
			if t.TagID == tagging.TagID && t.PostID == tagging.PostID {
				found = true
				break
			}
		}

		if !found {
			taggingsToCreate = append(taggingsToCreate, t)
		}
	}
	if len(taggingsToCreate) > 0 {
		newTaggings, err := CreateTaggings(ctx, db, taggingsToCreate)
		if err != nil {
			return results, errors.Wrap(err, "failed to create missing taggings")
		}

		results = append(results, newTaggings...)
	}

	return results, nil
}

func FindTaggingsByPostID(db *sql.DB, id int) (results []models.Tagging, err error) {
	var dbTaggings []dbTagging

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("photos.taggings").Select("*").Where(goqu.Ex{"post_id": id}).Executor()
	if err := insert.ScanStructs(&dbTaggings); err != nil {
		return results, errors.Wrap(err, "failed to select taggings by post_id")
	}

	for _, v := range dbTaggings {
		results = append(results, newTagging(v))
	}

	return results, nil
}

func FindTaggingsByTagID(ctx context.Context, db *sql.DB, id int) (results []models.Tagging, err error) {
	var dbTaggings []dbTagging

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("photos.taggings").Select("*").Where(goqu.Ex{"tag_id": id}).Executor()
	if err := insert.ScanStructsContext(ctx, &dbTaggings); err != nil {
		return results, errors.Wrap(err, "failed to select taggings by tag_id")
	}

	for _, v := range dbTaggings {
		results = append(results, newTagging(v))
	}

	return results, nil
}

func DeleteTaggings(ctx context.Context, db *sql.DB, taggings []models.Tagging) (err error) {
	var ids []int
	for _, d := range taggings {
		ids = append(ids, d.ID)
	}

	// nothing to delete
	if len(ids) == 0 {
		return nil
	}

	goquDB := goqu.New("postgres", db)
	del, _, err := goquDB.Delete("photos.taggings").Where(
		goqu.Ex{"id": ids},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("failed to build taggings delete query: %w", err)
	}
	_, err = db.ExecContext(ctx, del)
	if err != nil {
		return fmt.Errorf("failed to delete taggings: %w", err)
	}

	return nil
}

func AllTaggings(db *sql.DB) (results []models.Tagging, err error) {
	var dbTaggings []dbTagging

	goquDB := goqu.New("postgres", db)
	query := goquDB.From("photos.taggings").Select("*")

	if err := query.Executor().ScanStructs(&dbTaggings); err != nil {
		return results, errors.Wrap(err, "failed to select tags")
	}

	results = []models.Tagging{}
	for _, v := range dbTaggings {
		results = append(results, newTagging(v))
	}

	return results, nil
}
