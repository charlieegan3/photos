package database

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"

	"github.com/charlieegan3/photos/internal/pkg/models"
)

type dbPost struct {
	ID int `db:"id"`

	Description string `db:"description"`

	InstagramCode string `db:"instagram_code"`

	PublishDate time.Time `db:"publish_date"`

	IsDraft bool `db:"is_draft"`

	MediaID    int `db:"media_id"`
	LocationID int `db:"location_id"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (d *dbPost) ToRecord(includeID bool) goqu.Record {
	record := goqu.Record{
		"description":    d.Description,
		"instagram_code": d.InstagramCode,
		"is_draft":       d.IsDraft,
		"publish_date":   d.PublishDate.Format("2006-01-02 15:04:05"), // strip the zone since it's not in exif
		"media_id":       d.MediaID,
		"location_id":    d.LocationID,
	}

	if includeID {
		record["id"] = d.ID
	}

	return record
}

func newPost(post dbPost) models.Post {
	return models.Post{
		ID: post.ID,

		Description:   post.Description,
		InstagramCode: post.InstagramCode,
		PublishDate:   post.PublishDate.UTC(),

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

		Description:   post.Description,
		InstagramCode: post.InstagramCode,
		PublishDate:   post.PublishDate.UTC(),

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

func FindPostsByID(db *sql.DB, id []int) (results []models.Post, err error) {
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

func FindPostsByLocation(db *sql.DB, id []int) (results []models.Post, err error) {
	var dbPosts []dbPost

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("posts").Select("*").Where(goqu.Ex{"location_id": id}).Executor()
	if err := insert.ScanStructs(&dbPosts); err != nil {
		return results, errors.Wrap(err, "failed to select posts by id")
	}

	for _, v := range dbPosts {
		results = append(results, newPost(v))
	}

	return results, nil
}

func SearchPosts(db *sql.DB, query string) (results []models.Post, err error) {
	var dbPosts []dbPost

	safeQuery := regexp.MustCompile(`[^\w\s]+`).ReplaceAllString(query, "")
	matcher := regexp.MustCompile(fmt.Sprintf(`(^|\W)%s(s|\W|$)`, strings.ToLower(safeQuery)))

	goquDB := goqu.New("postgres", db)
	inner := goquDB.From("posts").
		Select("posts.*").
		Distinct("posts.id").
		LeftJoin(goqu.T("locations"), goqu.On(goqu.Ex{"posts.location_id": goqu.I("locations.id")})).
		LeftJoin(goqu.T("taggings"), goqu.On(goqu.Ex{"posts.id": goqu.I("taggings.post_id")})).
		LeftJoin(goqu.T("tags"), goqu.On(goqu.Ex{"taggings.tag_id": goqu.I("tags.id")})).
		GroupBy("posts.id", "locations.id", "tags.id").
		Having(
			goqu.Or(
				goqu.Ex{"posts.description": goqu.Op{"ilike": matcher}},
				goqu.Ex{"locations.name": goqu.Op{"ilike": matcher}},
				goqu.Ex{"tags.name": goqu.Op{"ilike": matcher}},
			),
		)
	outer := goquDB.From(inner).Select("*").Order(goqu.I("publish_date").Desc())
	if err := outer.ScanStructs(&dbPosts); err != nil {
		return results, errors.Wrap(err, "failed to select posts by id")
	}

	for _, v := range dbPosts {
		results = append(results, newPost(v))
	}

	return results, nil
}

func FindPostsByInstagramCode(db *sql.DB, code string) (results []models.Post, err error) {
	var dbPosts []dbPost

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("posts").Select("*").Where(goqu.Ex{"instagram_code": code}).Executor()
	if err := insert.ScanStructs(&dbPosts); err != nil {
		return results, errors.Wrap(err, "failed to select posts by instagram_code")
	}

	for _, v := range dbPosts {
		results = append(results, newPost(v))
	}

	return results, nil
}

func FindPostsByMediaID(db *sql.DB, id int) (results []models.Post, err error) {
	var dbPosts []dbPost

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("posts").Select("*").Where(goqu.Ex{"media_id": id}).Executor()
	if err := insert.ScanStructs(&dbPosts); err != nil {
		return results, errors.Wrap(err, "failed to select posts by media_id")
	}

	for _, v := range dbPosts {
		results = append(results, newPost(v))
	}

	return results, nil
}

func FindNextPost(db *sql.DB, post models.Post, previous bool) (results []models.Post, err error) {
	var dbPosts []dbPost

	query := goqu.C("publish_date").Gt(post.PublishDate)
	if previous {
		query = goqu.C("publish_date").Lt(post.PublishDate)
	}

	order := goqu.I("publish_date").Asc()
	if previous {
		order = goqu.I("publish_date").Desc()
	}

	goquDB := goqu.New("postgres", db)
	operation := goquDB.From("posts").Select("*").
		Where(query).
		Order(order).
		Limit(1).
		Executor()
	if err := operation.ScanStructs(&dbPosts); err != nil {
		return results, errors.Wrap(err, "failed to select posts by media_id")
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

func AllPosts(db *sql.DB, includeDrafts bool, options SelectOptions) (results []models.Post, err error) {
	var dbPosts []dbPost

	goquDB := goqu.New("postgres", db)
	query := goquDB.From("posts").Select("*")

	if !includeDrafts {
		query = query.Where(goqu.Ex{"is_draft": false})
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

	if err := query.Executor().ScanStructs(&dbPosts); err != nil {
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

func CountPosts(db *sql.DB, includeDrafts bool, options SelectOptions) (int64, error) {
	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("posts").Select("*")

	if !includeDrafts {
		insert = insert.Where(goqu.Ex{"is_draft": false})
	}

	if options.SortField != "" && options.SortDescending {
		insert = insert.Order(goqu.I(options.SortField).Desc())
	}

	if options.Offset != 0 {
		insert = insert.Offset(options.Offset)
	}

	if options.Limit != 0 {
		insert = insert.Limit(options.Limit)
	}

	count, err := insert.Count()
	if err != nil {
		return count, errors.Wrap(err, "failed to count posts")
	}

	return count, nil
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

func PostsInDateRange(db *sql.DB, after, before time.Time) (results []models.Post, err error) {
	var dbPosts []dbPost

	goquDB := goqu.New("postgres", db)
	query := goquDB.From("posts").
		Select("*").
		Where(
			goqu.And(
				goqu.Ex{
					"publish_date": goqu.Op{"gt": after},
				},
				goqu.Ex{
					"publish_date": goqu.Op{"lt": before},
				},
			),
		).
		Order(goqu.I("publish_date").Asc())

	if err := query.Executor().ScanStructs(&dbPosts); err != nil {
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

func PostsOnThisDay(db *sql.DB, month time.Month, day int) (results []models.Post, err error) {
	var dbPosts []dbPost

	goquDB := goqu.New("postgres", db)
	query := goquDB.From("posts").
		Select(
			"*",
		).
		Where(
			goqu.L(`EXTRACT(MONTH from publish_date)`).Eq(month),
			goqu.L(`EXTRACT(DAY from publish_date)`).Eq(day),
		).
		Order(goqu.I("publish_date").Desc())

	if err := query.Executor().ScanStructs(&dbPosts); err != nil {
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
