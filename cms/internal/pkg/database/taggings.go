package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"

	"github.com/charlieegan3/photos/cms/internal/pkg/models"
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

func CreateTaggings(db *sql.DB, taggings []models.Tagging) (results []models.Tagging, err error) {
	records := []goqu.Record{}
	for _, v := range taggings {
		d := newDBTagging(v)
		records = append(records, d.ToRecord(false))
	}

	var dbTaggings []dbTagging

	goquDB := goqu.New("postgres", db)
	insert := goquDB.Insert("taggings").Returning(goqu.Star()).Rows(records).Executor()
	if err := insert.ScanStructs(&dbTaggings); err != nil {
		return results, errors.Wrap(err, "failed to insert taggings")
	}

	for _, v := range dbTaggings {
		results = append(results, newTagging(v))
	}

	return results, nil
}

func FindTaggingsByPostID(db *sql.DB, id int) (results []models.Tagging, err error) {
	var dbTaggings []dbTagging

	goquDB := goqu.New("post", db)
	insert := goquDB.From("taggings").Select("*").Where(goqu.Ex{"post_id": id}).Executor()
	if err := insert.ScanStructs(&dbTaggings); err != nil {
		return results, errors.Wrap(err, "failed to select taggings by post_id")
	}

	for _, v := range dbTaggings {
		results = append(results, newTagging(v))
	}

	return results, nil
}

func DeleteTaggings(db *sql.DB, taggings []models.Tagging) (err error) {
	var ids []int
	for _, d := range taggings {
		ids = append(ids, d.ID)
	}

	goquDB := goqu.New("postgres", db)
	del, _, err := goquDB.Delete("taggings").Where(
		goqu.Ex{"id": ids},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("failed to build taggings delete query: %s", err)
	}
	_, err = db.Exec(del)
	if err != nil {
		return fmt.Errorf("failed to delete taggings: %s", err)
	}

	return nil
}
