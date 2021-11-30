package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"

	"github.com/charlieegan3/photos/cms/internal/pkg/models"
)

type dbPost struct {
	ID int `db:"id"`

	Description string `db:"description"`

	PublishDate time.Time `db:"publish_date"`

	IsDraft bool `db:"is_draft"`

	MediaID    int `db:"media_id"`
	LocationID int `db:"location_id"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (d *dbPost) ToRecord(includeID bool) goqu.Record {
	record := goqu.Record{
		"description":  d.Description,
		"is_draft":     d.IsDraft,
		"publish_date": d.PublishDate.Format("2006-01-02 15:04:05"), // strip the zone since it's not in exif
		"media_id":     d.MediaID,
		"location_id":  d.LocationID,
	}

	if includeID {
		record["id"] = d.ID
	}

	return record
}

func newPost(post dbPost) models.Post {
	return models.Post{
		ID: post.ID,

		Description: post.Description,
		PublishDate: post.PublishDate.UTC(),

		IsDraft: post.IsDraft,

		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,

		MediaID:    post.MediaID,
		LocationID: post.LocationID,
	}
}

func newDBPost(post models.Post) dbPost {
	return dbPost{
		ID: post.ID,

		Description: post.Description,
		PublishDate: post.PublishDate.UTC(),

		IsDraft: post.IsDraft,

		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,

		MediaID:    post.MediaID,
		LocationID: post.LocationID,
	}
}

func CreatePosts(db *sql.DB, posts []models.Post) (results []models.Post, err error) {
	records := []goqu.Record{}
	for _, v := range posts {
		d := newDBPost(v)
		records = append(records, d.ToRecord(false))
	}

	var dbPosts []dbPost

	goquDB := goqu.New("postgres", db)
	insert := goquDB.Insert("posts").Returning(goqu.Star()).Rows(records).Executor()
	if err := insert.ScanStructs(&dbPosts); err != nil {
		return results, errors.Wrap(err, "failed to insert posts")
	}

	for _, v := range dbPosts {
		results = append(results, newPost(v))
	}

	return results, nil
}

func FindPostsByID(db *sql.DB, id int) (results []models.Post, err error) {
	var dbPosts []dbPost

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("posts").Select("*").Where(goqu.Ex{"id": id}).Executor()
	if err := insert.ScanStructs(&dbPosts); err != nil {
		return results, errors.Wrap(err, "failed to select posts by id")
	}

	for _, v := range dbPosts {
		results = append(results, newPost(v))
	}

	return results, nil
}

func SetPostTags(db *sql.DB, post models.Post, rawTags []string) (err error) {
	var tags []models.Tag
	if len(rawTags) > 0 {
		tags, err = FindOrCreateTagsByName(db, rawTags)
		if err != nil {
			return errors.Wrap(err, "failed to find or created tags")
		}
	}

	existingTaggings, err := FindTaggingsByPostID(db, post.ID)
	if err != nil {
		return errors.Wrap(err, "failed to find existing taggings for post")
	}

	var requiredTaggings []models.Tagging
	for _, t := range tags {
		requiredTaggings = append(requiredTaggings, models.Tagging{
			PostID: post.ID,
			TagID:  t.ID,
		})
	}

	var taggingsToDelete []models.Tagging
	for _, tagging := range existingTaggings {
		found := false
		for _, t := range tags {
			if t.ID == tagging.TagID {
				found = true
				break
			}
		}

		if !found {
			taggingsToDelete = append(taggingsToDelete, tagging)
		}
	}

	err = DeleteTaggings(db, taggingsToDelete)
	if err != nil {
		return errors.Wrap(err, "failed to delete old taggings")
	}

	_, err = FindOrCreateTaggings(db, requiredTaggings)
	if err != nil {
		return errors.Wrap(err, "failed to find or create taggings")
	}

	return nil
}

func AllPosts(db *sql.DB) (results []models.Post, err error) {
	var dbPosts []dbPost

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("posts").Select("*").Executor()
	if err := insert.ScanStructs(&dbPosts); err != nil {
		return results, errors.Wrap(err, "failed to select posts")
	}

	// this is needed in case there are no items added, we don't want to return
	// nil but rather an empty slice
	results = []models.Post{}
	for _, v := range dbPosts {
		results = append(results, newPost(v))
	}

	return results, nil
}

func DeletePosts(db *sql.DB, posts []models.Post) (err error) {
	var ids []int
	for _, d := range posts {
		ids = append(ids, d.ID)
	}

	goquDB := goqu.New("postgres", db)
	del, _, err := goquDB.Delete("posts").Where(
		goqu.Ex{"id": ids},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("failed to build posts delete query: %s", err)
	}
	_, err = db.Exec(del)
	if err != nil {
		return fmt.Errorf("failed to delete posts: %s", err)
	}

	return nil
}

// UpdatePosts is not implemented as a single SQL query since update many in
// place is not supported by goqu and it wasn't worth the work (TODO)
func UpdatePosts(db *sql.DB, posts []models.Post) (results []models.Post, err error) {
	records := []goqu.Record{}
	for _, v := range posts {
		d := newDBPost(v)
		records = append(records, d.ToRecord(true))
	}

	goquDB := goqu.New("postgres", db)
	tx, err := goquDB.Begin()
	if err != nil {
		return results, errors.Wrap(err, "failed to open tx for updating posts")
	}

	for _, record := range records {
		var result dbPost
		update := tx.From("posts").
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

		results = append(results, newPost(result))
	}
	if err = tx.Commit(); err != nil {
		return results, errors.Wrap(err, "failed to commit transaction")
	}

	return results, nil
}
