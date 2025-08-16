package database

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/pkg/errors"

	"github.com/charlieegan3/photos/internal/pkg/models"
)

// PostFilterOptions provides filtering options for post queries.
type PostFilterOptions struct {
	Tags      []string
	Devices   []string
	Lenses    []string
	Locations []string
	Trips     []string

	From time.Time
	To   time.Time

	Limit  int
	Offset int
}

type dbPost struct {
	ID int `db:"id"`

	Description string `db:"description"`

	InstagramCode string `db:"instagram_code"`

	PublishDate time.Time `db:"publish_date"`

	IsDraft bool `db:"is_draft"`

	IsFavourite bool `db:"is_favourite"`

	MediaID    int `db:"media_id"`
	LocationID int `db:"location_id"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (d dbPost) ToRecord(includeID bool) goqu.Record {
	record := goqu.Record{
		"description":    d.Description,
		"instagram_code": d.InstagramCode,
		"is_draft":       d.IsDraft,
		"is_favourite":   d.IsFavourite,
		"publish_date":   d.PublishDate.Format("2006-01-02 15:04:05"), // strip the zone since it's not in exif
		"media_id":       d.MediaID,
		"location_id":    d.LocationID,
	}

	if includeID {
		record["id"] = d.ID
	}

	return record
}

func (d dbPost) ToModel() models.Post {
	return models.Post{
		ID: d.ID,

		Description:   d.Description,
		InstagramCode: d.InstagramCode,
		PublishDate:   d.PublishDate.UTC(),

		IsDraft:     d.IsDraft,
		IsFavourite: d.IsFavourite,

		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,

		MediaID:    d.MediaID,
		LocationID: d.LocationID,
	}
}

func newPost(post dbPost) models.Post {
	return post.ToModel()
}

func newDBPost(post models.Post) dbPost {
	return dbPost{
		ID: post.ID,

		Description:   post.Description,
		InstagramCode: post.InstagramCode,
		PublishDate:   post.PublishDate.UTC(),

		IsDraft:     post.IsDraft,
		IsFavourite: post.IsFavourite,

		CreatedAt: post.CreatedAt,
		UpdatedAt: post.UpdatedAt,

		MediaID:    post.MediaID,
		LocationID: post.LocationID,
	}
}

// PostRepository provides post-specific database operations.
type PostRepository struct {
	*BaseRepository[models.Post, dbPost]
}

// NewPostRepository creates a new post repository instance.
func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{
		BaseRepository: NewBaseRepository(db, "posts", newPost, newDBPost, "publish_date"),
	}
}

// RandomPostID returns a random post ID.
func (r *PostRepository) RandomPostID(ctx context.Context) (int, error) {
	var result struct {
		ID int `db:"id"`
	}

	sql := `SELECT id FROM photos.posts WHERE is_draft = false ORDER BY RANDOM() LIMIT 1`
	err := r.db.QueryRowContext(ctx, sql).Scan(&result.ID)
	if err != nil {
		return 0, errors.Wrap(err, "failed to select random post")
	}

	return result.ID, nil
}

// FindByLocation finds posts by location IDs.
func (r *PostRepository) FindByLocation(ctx context.Context, locationIDs []int) ([]models.Post, error) {
	var dbPosts []dbPost

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select("*").
		Where(goqu.Ex{"location_id": locationIDs}).
		Order(goqu.I("publish_date").Desc()).
		Executor()

	err := query.ScanStructsContext(ctx, &dbPosts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select posts by location")
	}

	results := make([]models.Post, 0, len(dbPosts))
	for i := range dbPosts {
		results = append(results, newPost(dbPosts[i]))
	}

	return results, nil
}

// Search performs text search on posts.
func (r *PostRepository) Search(ctx context.Context, query string) ([]models.Post, error) {
	if query == "" {
		return []models.Post{}, nil
	}

	// Sanitize query for SQL ILIKE pattern
	safeQuery := strings.TrimSpace(query)
	safeQuery = strings.ReplaceAll(safeQuery, "%", "\\%")
	safeQuery = strings.ReplaceAll(safeQuery, "_", "\\_")
	searchPattern := "%" + safeQuery + "%"

	var dbPosts []dbPost
	goquDB := goqu.New("postgres", r.db)

	// Search in descriptions using ILIKE
	descQuery := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select("*").
		Where(goqu.L("description ILIKE ?", searchPattern)).
		Order(goqu.I("publish_date").Desc()).
		Executor()

	err := descQuery.ScanStructsContext(ctx, &dbPosts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search posts by description")
	}

	// Convert results
	results := make([]models.Post, 0, len(dbPosts))
	for i := range dbPosts {
		post := newPost(dbPosts[i])
		results = append(results, post)
	}

	// Also search in tags
	var tagPosts []dbPost
	tagQuery := goquDB.From(goqu.T("tags").Schema("photos")).
		InnerJoin(goqu.T("taggings").Schema("photos"), goqu.On(goqu.Ex{"taggings.tag_id": goqu.I("tags.id")})).
		InnerJoin(goqu.T("posts").Schema("photos"), goqu.On(goqu.Ex{"posts.id": goqu.I("taggings.post_id")})).
		Select("posts.*").
		Where(goqu.L(`tags.name ILIKE ?`, searchPattern)).
		Order(goqu.I("posts.publish_date").Desc()).
		Executor()

	err = tagQuery.ScanStructsContext(ctx, &tagPosts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search posts by tags")
	}

	// Also search in location names
	var locationPosts []dbPost
	locationQuery := goquDB.From(goqu.T("locations").Schema("photos")).
		InnerJoin(goqu.T("posts").Schema("photos"), goqu.On(goqu.Ex{"posts.location_id": goqu.I("locations.id")})).
		Select("posts.*").
		Where(goqu.L(`locations.name ILIKE ?`, searchPattern)).
		Order(goqu.I("posts.publish_date").Desc()).
		Executor()

	err = locationQuery.ScanStructsContext(ctx, &locationPosts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search posts by location")
	}

	// Merge results and deduplicate
	postMap := make(map[int]models.Post)
	for i := range results {
		postMap[results[i].ID] = results[i]
	}
	for i := range tagPosts {
		post := newPost(tagPosts[i])
		postMap[post.ID] = post
	}
	for i := range locationPosts {
		post := newPost(locationPosts[i])
		postMap[post.ID] = post
	}

	// Convert map back to slice
	results = make([]models.Post, 0, len(postMap))
	for postID := range postMap {
		results = append(results, postMap[postID])
	}

	// Sort by publish date descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].PublishDate.After(results[j].PublishDate)
	})

	return results, nil
}

// FindByInstagramCode finds posts by Instagram code.
func (r *PostRepository) FindByInstagramCode(ctx context.Context, code string) ([]models.Post, error) {
	return r.FindByField(ctx, "instagram_code", code)
}

// FindByMediaID finds posts by media ID.
func (r *PostRepository) FindByMediaID(ctx context.Context, mediaID int) ([]models.Post, error) {
	return r.FindByField(ctx, "media_id", mediaID)
}

// FindNextPost finds the next or previous post relative to a given post.
func (r *PostRepository) FindNextPost(post models.Post, previous bool) ([]models.Post, error) {
	var dbPosts []dbPost

	goquDB := goqu.New("postgres", r.db)

	var operator string
	var order exp.OrderedExpression
	if previous {
		operator = "<"
		order = goqu.I("publish_date").Desc()
	} else {
		operator = ">"
		order = goqu.I("publish_date").Asc()
	}

	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select("*").
		Where(goqu.L(fmt.Sprintf(`"publish_date" %s ?`, operator), post.PublishDate)).
		Order(order).
		Limit(1).
		Executor()

	err := query.ScanStructsContext(context.Background(), &dbPosts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find next post")
	}

	results := make([]models.Post, 0, len(dbPosts))
	for i := range dbPosts {
		results = append(results, newPost(dbPosts[i]))
	}

	return results, nil
}

// SetTags sets tags for a post.
func (r *PostRepository) SetTags(ctx context.Context, post models.Post, rawTags []string) error {
	// Filter out empty tags
	var filteredTags []string
	for _, tag := range rawTags {
		if strings.TrimSpace(tag) != "" {
			filteredTags = append(filteredTags, strings.TrimSpace(tag))
		}
	}

	// First, find or create the tags
	var tags []models.Tag
	var err error
	if len(filteredTags) > 0 {
		tags, err = FindOrCreateTagsByName(ctx, r.db, filteredTags)
		if err != nil {
			return errors.Wrap(err, "failed to find or create tags")
		}
	}

	goquDB := goqu.New("postgres", r.db)
	tx, err := goquDB.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Delete existing taggings
	_, err = tx.Delete("photos.taggings").
		Where(goqu.Ex{"post_id": post.ID}).
		Executor().
		ExecContext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to delete existing taggings")
	}

	// Insert new taggings
	for _, tag := range tags {
		_, err = tx.Insert("photos.taggings").
			Rows(goqu.Record{"post_id": post.ID, "tag_id": tag.ID}).
			Executor().
			ExecContext(ctx)
		if err != nil {
			return errors.Wrapf(err, "failed to create tagging for tag '%s'", tag.Name)
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

// AllWithOptions retrieves all posts with filtering options.
func (r *PostRepository) AllWithOptions(
	ctx context.Context, includeDrafts bool, options PostFilterOptions,
) ([]models.Post, error) {
	var dbPosts []dbPost

	goquDB := goqu.New("postgres", r.db)
	query := r.buildBaseQueryWithJoins(goquDB).Select("posts.*")

	// Add filtering conditions
	conditions := goqu.Ex{}
	if !includeDrafts {
		conditions["posts.is_draft"] = false
	}

	if len(options.Tags) > 0 {
		query = query.InnerJoin(
			goqu.T("taggings").Schema("photos"),
			goqu.On(goqu.Ex{"taggings.post_id": goqu.I("posts.id")}),
		).
			InnerJoin(goqu.T("tags").Schema("photos"), goqu.On(goqu.Ex{"tags.id": goqu.I("taggings.tag_id")}))
		conditions["tags.name"] = options.Tags
	}

	if len(options.Devices) > 0 {
		conditions["devices.name"] = options.Devices
	}
	if len(options.Lenses) > 0 {
		conditions["lenses.name"] = options.Lenses
	}
	if len(options.Locations) > 0 {
		conditions["locations.name"] = options.Locations
	}
	if len(options.Trips) > 0 {
		conditions["trips.title"] = options.Trips
	}

	if !options.From.IsZero() {
		conditions["posts.publish_date"] = goqu.Op{"gte": options.From}
	}
	if !options.To.IsZero() {
		conditions["posts.publish_date"] = goqu.Op{"lte": options.To}
	}

	query = query.Where(conditions).
		GroupBy(goqu.I("posts.id")).
		Order(goqu.I("posts.publish_date").Desc())

	if options.Limit > 0 {
		query = query.Limit(uint(options.Limit))
	}
	if options.Offset > 0 {
		query = query.Offset(uint(options.Offset))
	}

	err := query.Executor().ScanStructsContext(ctx, &dbPosts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select posts with options")
	}

	results := make([]models.Post, 0, len(dbPosts))
	for i := range dbPosts {
		results = append(results, newPost(dbPosts[i]))
	}

	return results, nil
}

// Count returns the count of posts with filtering options.
func (r *PostRepository) Count(ctx context.Context, includeDrafts bool, options PostFilterOptions) (uint, error) {
	goquDB := goqu.New("postgres", r.db)
	query := r.buildBaseQueryWithJoins(goquDB).Select(goqu.COUNT(goqu.DISTINCT("posts.id")))

	// Add same filtering logic as AllWithOptions
	conditions := goqu.Ex{}
	if !includeDrafts {
		conditions["posts.is_draft"] = false
	}

	if len(options.Tags) > 0 {
		query = query.InnerJoin(
			goqu.T("taggings").Schema("photos"),
			goqu.On(goqu.Ex{"taggings.post_id": goqu.I("posts.id")}),
		).
			InnerJoin(goqu.T("tags").Schema("photos"), goqu.On(goqu.Ex{"tags.id": goqu.I("taggings.tag_id")}))
		conditions["tags.name"] = options.Tags
	}

	if len(options.Devices) > 0 {
		conditions["devices.name"] = options.Devices
	}
	if len(options.Lenses) > 0 {
		conditions["lenses.name"] = options.Lenses
	}
	if len(options.Locations) > 0 {
		conditions["locations.name"] = options.Locations
	}
	if len(options.Trips) > 0 {
		conditions["trips.title"] = options.Trips
	}

	if !options.From.IsZero() {
		conditions["posts.publish_date"] = goqu.Op{"gte": options.From}
	}
	if !options.To.IsZero() {
		conditions["posts.publish_date"] = goqu.Op{"lte": options.To}
	}

	query = query.Where(conditions)

	var count uint
	sql, args, err := query.ToSQL()
	if err != nil {
		return 0, errors.Wrap(err, "failed to build count query")
	}
	err = r.db.QueryRowContext(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "failed to count posts")
	}

	return count, nil
}

// InDateRange finds posts within a date range.
func (r *PostRepository) InDateRange(ctx context.Context, after, before time.Time) ([]models.Post, error) {
	var dbPosts []dbPost

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select("*").
		Where(
			goqu.And(
				goqu.I("publish_date").Gte(after),
				goqu.I("publish_date").Lte(before),
				goqu.I("is_draft").Eq(false),
			),
		).
		Order(goqu.I("publish_date").Desc()).
		Executor()

	err := query.ScanStructsContext(ctx, &dbPosts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select posts in date range")
	}

	results := make([]models.Post, 0, len(dbPosts))
	for i := range dbPosts {
		results = append(results, newPost(dbPosts[i]))
	}

	return results, nil
}

// OnThisDay finds posts on this day (month/day) from previous years.
func (r *PostRepository) OnThisDay(ctx context.Context, month time.Month, day int) ([]models.Post, error) {
	var dbPosts []dbPost

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select("*").
		Where(goqu.L(
			"EXTRACT(month FROM publish_date) = ? AND EXTRACT(day FROM publish_date) = ? AND is_draft = false",
			int(month), day,
		)).
		Order(goqu.I("publish_date").Desc()).
		Executor()

	err := query.ScanStructsContext(ctx, &dbPosts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select posts on this day")
	}

	results := make([]models.Post, 0, len(dbPosts))
	for i := range dbPosts {
		results = append(results, newPost(dbPosts[i]))
	}

	return results, nil
}

// buildBaseQueryWithJoins creates the base query with all common joins.
func (r *PostRepository) buildBaseQueryWithJoins(goquDB *goqu.Database) *goqu.SelectDataset {
	return goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		InnerJoin(goqu.T("medias").Schema("photos"), goqu.On(goqu.Ex{"medias.id": goqu.I("posts.media_id")})).
		LeftJoin(goqu.T("devices").Schema("photos"), goqu.On(goqu.Ex{"devices.id": goqu.I("medias.device_id")})).
		LeftJoin(goqu.T("lenses").Schema("photos"), goqu.On(goqu.Ex{"lenses.id": goqu.I("medias.lens_id")})).
		LeftJoin(goqu.T("locations").Schema("photos"), goqu.On(goqu.Ex{"locations.id": goqu.I("posts.location_id")})).
		LeftJoin(goqu.T("trips").Schema("photos"), goqu.On(goqu.Ex{"trips.id": goqu.I("medias.trip_id")}))
}

// Legacy function wrappers for backward compatibility with test files.
// These should be removed after all tests are updated.

// CreatePosts creates multiple posts using the repository.
func CreatePosts(ctx context.Context, db *sql.DB, posts []models.Post) ([]models.Post, error) {
	repo := NewPostRepository(db)
	return repo.Create(ctx, posts)
}

// RandomPostID returns a random post ID.
func RandomPostID(ctx context.Context, db *sql.DB) (int, error) {
	repo := NewPostRepository(db)
	return repo.RandomPostID(ctx)
}

// FindPostsByID finds posts by their IDs using the repository.
func FindPostsByID(ctx context.Context, db *sql.DB, ids []int) ([]models.Post, error) {
	repo := NewPostRepository(db)
	int64IDs := make([]int64, len(ids))
	for i, id := range ids {
		int64IDs[i] = int64(id)
	}
	return repo.FindByIDs(ctx, int64IDs)
}

// FindPostsByLocation finds posts by location IDs.
func FindPostsByLocation(ctx context.Context, db *sql.DB, locationIDs []int) ([]models.Post, error) {
	repo := NewPostRepository(db)
	return repo.FindByLocation(ctx, locationIDs)
}

// SearchPosts performs text search on posts.
func SearchPosts(ctx context.Context, db *sql.DB, query string) ([]models.Post, error) {
	repo := NewPostRepository(db)
	return repo.Search(ctx, query)
}

// FindPostsByInstagramCode finds posts by Instagram code.
func FindPostsByInstagramCode(ctx context.Context, db *sql.DB, code string) ([]models.Post, error) {
	repo := NewPostRepository(db)
	return repo.FindByInstagramCode(ctx, code)
}

// FindPostsByMediaID finds posts by media ID.
func FindPostsByMediaID(ctx context.Context, db *sql.DB, mediaID int) ([]models.Post, error) {
	repo := NewPostRepository(db)
	return repo.FindByMediaID(ctx, mediaID)
}

// FindNextPost finds the next or previous post.
func FindNextPost(db *sql.DB, post models.Post, previous bool) ([]models.Post, error) {
	repo := NewPostRepository(db)
	return repo.FindNextPost(post, previous)
}

// SetPostTags sets tags for a post.
func SetPostTags(ctx context.Context, db *sql.DB, post models.Post, rawTags []string) error {
	repo := NewPostRepository(db)
	return repo.SetTags(ctx, post, rawTags)
}

// FavouritePosts retrieves all favourite posts ordered by created_at descending.
func FavouritePosts(ctx context.Context, db *sql.DB) ([]models.Post, error) {
	var dbPosts []dbPost

	goquDB := goqu.New("postgres", db)
	query := goquDB.From("photos.posts").Select("*").
		Where(goqu.Ex{"is_draft": false, "is_favourite": true}).
		Order(goqu.I("created_at").Desc())

	err := query.Executor().ScanStructsContext(ctx, &dbPosts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select favourite posts")
	}

	posts := make([]models.Post, len(dbPosts))
	for i := range dbPosts {
		posts[i] = dbPosts[i].ToModel()
	}

	return posts, nil
}

// AllPosts retrieves all posts with basic sorting and pagination (original behavior).
func AllPosts(ctx context.Context, db *sql.DB, includeDrafts bool, options SelectOptions) ([]models.Post, error) {
	var dbPosts []dbPost

	goquDB := goqu.New("postgres", db)
	query := goquDB.From("photos.posts").Select("*")

	if !includeDrafts {
		query = query.Where(goqu.Ex{"is_draft": false})
	}

	if options.SortField != "" {
		if options.SortDescending {
			query = query.Order(goqu.I(options.SortField).Desc())
		} else {
			query = query.Order(goqu.I(options.SortField).Asc())
		}
	}

	if options.Offset != 0 {
		query = query.Offset(options.Offset)
	}

	if options.Limit != 0 {
		query = query.Limit(options.Limit)
	}

	err := query.Executor().ScanStructsContext(ctx, &dbPosts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select posts")
	}

	results := make([]models.Post, 0, len(dbPosts))
	for i := range dbPosts {
		results = append(results, newPost(dbPosts[i]))
	}

	return results, nil
}

// CountPosts returns the count of posts (original behavior).
func CountPosts(ctx context.Context, db *sql.DB, includeDrafts bool, _ SelectOptions) (uint, error) {
	goquDB := goqu.New("postgres", db)
	query := goquDB.From("photos.posts").Select(goqu.COUNT("*"))

	if !includeDrafts {
		query = query.Where(goqu.Ex{"is_draft": false})
	}

	var count uint
	found, err := query.Executor().ScanValContext(ctx, &count)
	if err != nil {
		return 0, errors.Wrap(err, "failed to count posts")
	}
	if !found {
		return 0, nil
	}

	return count, nil
}

// DeletePosts deletes posts using the repository.
func DeletePosts(ctx context.Context, db *sql.DB, posts []models.Post) error {
	repo := NewPostRepository(db)
	return repo.Delete(ctx, posts)
}

// UpdatePosts updates posts using the repository.
func UpdatePosts(ctx context.Context, db *sql.DB, posts []models.Post) ([]models.Post, error) {
	repo := NewPostRepository(db)
	return repo.Update(ctx, posts)
}

// PostsInDateRange finds posts within a date range.
func PostsInDateRange(ctx context.Context, db *sql.DB, after, before time.Time) ([]models.Post, error) {
	repo := NewPostRepository(db)
	return repo.InDateRange(ctx, after, before)
}

// PostsOnThisDay finds posts on this day from previous years.
func PostsOnThisDay(ctx context.Context, db *sql.DB, month time.Month, day int) ([]models.Post, error) {
	repo := NewPostRepository(db)
	return repo.OnThisDay(ctx, month, day)
}
