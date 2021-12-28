package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"

	"github.com/charlieegan3/photos/cms/internal/pkg/models"
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

type TagNameConflictError struct{}

func (t TagNameConflictError) Error() string {
	return "tag name conflicted with an existing tag"
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

func FindTagsByName(db *sql.DB, names []string) (results []models.Tag, err error) {
	var dbTags []dbTag

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("tags").Select("*").Where(goqu.I("name").In(names)).Executor()
	if err := insert.ScanStructs(&dbTags); err != nil {
		return results, errors.Wrap(err, "failed to select tags by name")
	}

	for _, v := range dbTags {
		results = append(results, newTag(v))
	}

	return results, nil
}

func FindTagsByID(db *sql.DB, ids []int) (results []models.Tag, err error) {
	if len(ids) == 0 {
		return results, nil
	}

	var dbTags []dbTag

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("tags").Select("*").Where(goqu.I("id").In(ids)).Executor()
	if err := insert.ScanStructs(&dbTags); err != nil {
		return results, errors.Wrap(err, "failed to select tags by id")
	}

	for _, v := range dbTags {
		results = append(results, newTag(v))
	}

	return results, nil
}

func FindOrCreateTagsByName(db *sql.DB, names []string) (results []models.Tag, err error) {
	resultMap := make(map[string]models.Tag)

	foundTags, err := FindTagsByName(db, names)
	for _, t := range foundTags {
		resultMap[t.Name] = t
	}

	var tagsToCreate []models.Tag
	for _, n := range names {
		if _, ok := resultMap[n]; !ok {
			tagsToCreate = append(tagsToCreate, models.Tag{Name: n})
		}
	}

	createdTags, err := CreateTags(db, tagsToCreate)
	for _, t := range createdTags {
		resultMap[t.Name] = t
	}

	for _, t := range names {
		tag, ok := resultMap[t]
		if !ok {
			return results, fmt.Errorf("expected tag %q to have been found or created", t)
		}

		results = append(results, tag)
	}

	return results, nil
}

func AllTags(db *sql.DB, includeHidden bool, options SelectOptions) (results []models.Tag, err error) {
	var dbTags []dbTag

	goquDB := goqu.New("postgres", db)
	query := goquDB.From("tags").Select("*")

	if !includeHidden {
		query = query.Where(goqu.Ex{"hidden": false})
	}

	if options.SortField != "" {
		query = query.Order(goqu.I(options.SortField).Asc())
	}
	if options.SortField != "" && options.SortDescending {
		query = query.Order(goqu.I(options.SortField).Desc())
	}

	if options.Offset != 0 {
		query = query.Offset(options.Offset)
	}

	if options.Limit != 0 {
		query = query.Limit(options.Limit)
	}

	if err := query.Executor().ScanStructs(&dbTags); err != nil {
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

			return results, &TagNameConflictError{}
		}

		results = append(results, newTag(result))
	}
	if err = tx.Commit(); err != nil {
		return results, errors.Wrap(err, "failed to commit transaction")
	}

	return results, nil
}

// TODO make this a transaction
func MergeTags(db *sql.DB, tag1, tag2 models.Tag) (err error) {
	taggings, err := FindTaggingsByTagID(db, tag2.ID)
	if err != nil {
		return errors.Wrap(err, "failed to list all taggings")
	}

	// create new taggings for all posts to tag1
	var newTaggings []models.Tagging
	for _, t := range taggings {
		newTaggings = append(newTaggings, models.Tagging{TagID: tag1.ID, PostID: t.PostID})
	}

	_, err = CreateTaggings(db, newTaggings)
	if err != nil {
		return errors.Wrap(err, "failed to create new taggings for tag1")
	}

	err = DeleteTags(db, []models.Tag{tag2})
	if err != nil {
		return errors.Wrap(err, "failed to delete tag2")
	}

	return nil
}
