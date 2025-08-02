package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"

	"github.com/charlieegan3/photos/internal/pkg/models"
)

type dbTag struct {
	ID     int    `db:"id"`
	Name   string `db:"name"`
	Hidden bool   `db:"hidden"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (d dbTag) ToRecord(includeID bool) goqu.Record {
	record := goqu.Record{
		"name":   d.Name,
		"hidden": d.Hidden,
	}

	if includeID {
		record["id"] = d.ID
	}

	return record
}

func (d dbTag) ToModel() models.Tag {
	return models.Tag{
		ID:        d.ID,
		Name:      d.Name,
		Hidden:    d.Hidden,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func newTag(tag dbTag) models.Tag {
	return tag.ToModel()
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

func (TagNameConflictError) Error() string {
	return "tag name conflicted with an existing tag"
}

// TagRepository provides tag-specific database operations.
type TagRepository struct {
	*BaseRepository[models.Tag, dbTag]
}

// NewTagRepository creates a new tag repository instance.
func NewTagRepository(db *sql.DB) *TagRepository {
	return &TagRepository{
		BaseRepository: NewBaseRepository(db, "tags", newTag, newDBTag, "created_at"),
	}
}

// FindByName finds tags by their names.
func (r *TagRepository) FindByName(ctx context.Context, names []string) ([]models.Tag, error) {
	var dbTags []dbTag

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select("*").
		Where(goqu.I("name").In(names)).
		Executor()

	err := query.ScanStructsContext(ctx, &dbTags)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select tags by name")
	}

	results := make([]models.Tag, 0, len(dbTags))
	for _, tag := range dbTags {
		results = append(results, newTag(tag))
	}

	return results, nil
}

// FindByIntIDs finds tags by their int IDs (compatibility wrapper).
func (r *TagRepository) FindByIntIDs(ctx context.Context, ids []int) ([]models.Tag, error) {
	if len(ids) == 0 {
		return []models.Tag{}, nil
	}

	var dbTags []dbTag

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select("*").
		Where(goqu.I("id").In(ids)).
		Executor()

	err := query.ScanStructsContext(ctx, &dbTags)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select tags by id")
	}

	results := make([]models.Tag, 0, len(dbTags))
	for _, tag := range dbTags {
		results = append(results, newTag(tag))
	}

	return results, nil
}

// FindOrCreateByName finds or creates tags by their names.
func (r *TagRepository) FindOrCreateByName(ctx context.Context, names []string) ([]models.Tag, error) {
	resultMap := make(map[string]models.Tag)

	foundTags, err := r.FindByName(ctx, names)
	if err != nil {
		return nil, err
	}
	for _, t := range foundTags {
		resultMap[t.Name] = t
	}

	var tagsToCreate []models.Tag
	for _, n := range names {
		if _, ok := resultMap[n]; !ok {
			tagsToCreate = append(tagsToCreate, models.Tag{Name: n})
		}
	}

	if len(tagsToCreate) > 0 {
		createdTags, err := r.Create(ctx, tagsToCreate)
		if err != nil {
			return nil, err
		}
		for _, t := range createdTags {
			resultMap[t.Name] = t
		}
	}

	results := make([]models.Tag, 0, len(names))
	for _, t := range names {
		tag, ok := resultMap[t]
		if !ok {
			return nil, fmt.Errorf("expected tag %q to have been found or created", t)
		}

		results = append(results, tag)
	}

	return results, nil
}

// AllWithOptions retrieves all tags with additional filtering and sorting options.
func (r *TagRepository) AllWithOptions(
	ctx context.Context, includeHidden bool, options SelectOptions,
) ([]models.Tag, error) {
	var dbTags []dbTag

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).Select("*")

	if !includeHidden {
		query = query.Where(goqu.Ex{"hidden": false})
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

	err := query.Executor().ScanStructsContext(ctx, &dbTags)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select tags")
	}

	// this is needed in case there are no items added, we don't want to return
	// nil but rather an empty slice
	results := make([]models.Tag, 0, len(dbTags))
	for _, tag := range dbTags {
		results = append(results, newTag(tag))
	}

	return results, nil
}

// Merge merges two tags by moving all taggings from tag2 to tag1, then deleting tag2.
func (r *TagRepository) Merge(ctx context.Context, tag1, tag2 models.Tag) error {
	taggings, err := FindTaggingsByTagID(ctx, r.db, tag2.ID)
	if err != nil {
		return errors.Wrap(err, "failed to list all taggings")
	}

	if len(taggings) > 0 {
		// create new taggings for all posts to tag1
		var newTaggings []models.Tagging
		for _, t := range taggings {
			newTaggings = append(newTaggings, models.Tagging{TagID: tag1.ID, PostID: t.PostID})
		}

		_, err = CreateTaggings(ctx, r.db, newTaggings)
		if err != nil {
			return errors.Wrap(err, "failed to create new taggings for tag1")
		}
	}

	err = r.Delete(ctx, []models.Tag{tag2})
	if err != nil {
		return errors.Wrap(err, "failed to delete tag2")
	}

	return nil
}

// Legacy function wrappers for backward compatibility with test files.
// These should be removed after all tests are updated.

// CreateTags creates multiple tags using the repository.
func CreateTags(ctx context.Context, db *sql.DB, tags []models.Tag) ([]models.Tag, error) {
	repo := NewTagRepository(db)
	return repo.Create(ctx, tags)
}

// FindTagsByName finds tags by name using the repository.
func FindTagsByName(ctx context.Context, db *sql.DB, names []string) ([]models.Tag, error) {
	repo := NewTagRepository(db)
	return repo.FindByName(ctx, names)
}

// FindTagsByID finds tags by their IDs using the repository.
func FindTagsByID(ctx context.Context, db *sql.DB, ids []int) ([]models.Tag, error) {
	repo := NewTagRepository(db)
	return repo.FindByIntIDs(ctx, ids)
}

// FindOrCreateTagsByName finds or creates tags by name using the repository.
func FindOrCreateTagsByName(ctx context.Context, db *sql.DB, names []string) ([]models.Tag, error) {
	repo := NewTagRepository(db)
	return repo.FindOrCreateByName(ctx, names)
}

// AllTags gets all tags using the repository.
func AllTags(ctx context.Context, db *sql.DB, includeHidden bool, options SelectOptions) ([]models.Tag, error) {
	repo := NewTagRepository(db)
	return repo.AllWithOptions(ctx, includeHidden, options)
}

// DeleteTags deletes tags using the repository.
func DeleteTags(ctx context.Context, db *sql.DB, tags []models.Tag) error {
	if len(tags) == 0 {
		return nil
	}

	ids := make([]int, len(tags))
	for i, tag := range tags {
		ids[i] = tag.ID
	}

	goquDB := goqu.New("postgres", db)
	query := goquDB.From("photos.tags").Where(goqu.Ex{"id": ids}).Delete()
	_, err := query.Executor().ExecContext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to delete tags")
	}

	return nil
}

// UpdateTags updates tags using the repository.
func UpdateTags(ctx context.Context, db *sql.DB, tags []models.Tag) ([]models.Tag, error) {
	results, err := BulkUpdateGeneric(ctx, db, "photos.tags", tags, newDBTag)
	if err != nil {
		return results, &TagNameConflictError{}
	}
	return results, nil
}

// MergeTags merges two tags using the repository.
func MergeTags(ctx context.Context, db *sql.DB, tag1, tag2 models.Tag) error {
	repo := NewTagRepository(db)
	return repo.Merge(ctx, tag1, tag2)
}
