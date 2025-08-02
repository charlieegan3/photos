package database

import (
	"context"
	"database/sql"
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

func (d dbTagging) ToRecord(includeID bool) goqu.Record {
	record := goqu.Record{
		"post_id": d.PostID,
		"tag_id":  d.TagID,
	}

	if includeID {
		record["id"] = d.ID
	}

	return record
}

func (d dbTagging) ToModel() models.Tagging {
	return models.Tagging{
		ID: d.ID,

		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,

		PostID: d.PostID,
		TagID:  d.TagID,
	}
}

func newTagging(tagging dbTagging) models.Tagging {
	return tagging.ToModel()
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

// TaggingRepository provides tagging-specific database operations.
type TaggingRepository struct {
	*BaseRepository[models.Tagging, dbTagging]
}

// NewTaggingRepository creates a new tagging repository instance.
func NewTaggingRepository(db *sql.DB) *TaggingRepository {
	return &TaggingRepository{
		BaseRepository: NewBaseRepository(db, "taggings", newTagging, newDBTagging, "post_id"),
	}
}

// All retrieves all taggings ordered by post_id ASC for predictable ordering.
func (r *TaggingRepository) All(ctx context.Context) ([]models.Tagging, error) {
	var dbTaggings []dbTagging

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select("*").
		Order(goqu.I("post_id").Asc()).
		Executor()

	err := query.ScanStructsContext(ctx, &dbTaggings)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select all taggings")
	}

	results := make([]models.Tagging, 0, len(dbTaggings))
	for _, dbTagging := range dbTaggings {
		results = append(results, newTagging(dbTagging))
	}

	return results, nil
}

// FindByPostID finds taggings by post ID.
func (r *TaggingRepository) FindByPostID(ctx context.Context, id int) ([]models.Tagging, error) {
	var dbTaggings []dbTagging

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select("*").
		Where(goqu.Ex{"post_id": id}).
		Executor()

	err := query.ScanStructsContext(ctx, &dbTaggings)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select taggings by post_id")
	}

	results := make([]models.Tagging, 0, len(dbTaggings))
	for _, tagging := range dbTaggings {
		results = append(results, newTagging(tagging))
	}

	return results, nil
}

// FindByTagID finds taggings by tag ID.
func (r *TaggingRepository) FindByTagID(ctx context.Context, id int) ([]models.Tagging, error) {
	var dbTaggings []dbTagging

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select("*").
		Where(goqu.Ex{"tag_id": id}).
		Executor()

	err := query.ScanStructsContext(ctx, &dbTaggings)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select taggings by tag_id")
	}

	results := make([]models.Tagging, 0, len(dbTaggings))
	for _, tagging := range dbTaggings {
		results = append(results, newTagging(tagging))
	}

	return results, nil
}

// CreateWithConflictHandling creates taggings with ON CONFLICT DO NOTHING behavior.
func (r *TaggingRepository) CreateWithConflictHandling(
	ctx context.Context, taggings []models.Tagging,
) ([]models.Tagging, error) {
	if len(taggings) == 0 {
		return []models.Tagging{}, nil
	}

	records := make([]goqu.Record, 0, len(taggings))
	for _, tagging := range taggings {
		dbTagging := newDBTagging(tagging)
		records = append(records, dbTagging.ToRecord(false))
	}

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.Insert(goqu.T(r.tableName).Schema(r.schema)).
		Returning(goqu.Star()).
		Rows(records).
		OnConflict(goqu.DoNothing()).
		Executor()

	var dbTaggings []dbTagging
	err := query.ScanStructsContext(ctx, &dbTaggings)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert taggings")
	}

	results := make([]models.Tagging, 0, len(dbTaggings))
	for _, dbTagging := range dbTaggings {
		results = append(results, newTagging(dbTagging))
	}

	return results, nil
}

// FindOrCreate finds existing taggings or creates missing ones.
func (r *TaggingRepository) FindOrCreate(ctx context.Context, taggings []models.Tagging) ([]models.Tagging, error) {
	ex := make([]exp.Expression, len(taggings))
	for i, t := range taggings {
		ex[i] = goqu.Ex{
			"post_id": t.PostID,
			"tag_id":  t.TagID,
		}
	}

	var dbTaggings []dbTagging

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select("*").
		Where(goqu.Or(ex...)).
		Executor()

	err := query.ScanStructsContext(ctx, &dbTaggings)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select taggings by post_id")
	}

	results := make([]models.Tagging, 0, len(dbTaggings))
	for _, tagging := range dbTaggings {
		results = append(results, newTagging(tagging))
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
		newTaggings, err := r.CreateWithConflictHandling(ctx, taggingsToCreate)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create missing taggings")
		}

		results = append(results, newTaggings...)
	}

	return results, nil
}

// Legacy function wrappers for backward compatibility with test files.
// These should be removed after all tests are updated.

// CreateTaggings creates multiple taggings using the repository.
func CreateTaggings(ctx context.Context, db *sql.DB, taggings []models.Tagging) ([]models.Tagging, error) {
	repo := NewTaggingRepository(db)
	return repo.CreateWithConflictHandling(ctx, taggings)
}

// FindOrCreateTaggings finds or creates taggings using the repository.
func FindOrCreateTaggings(ctx context.Context, db *sql.DB, taggings []models.Tagging) ([]models.Tagging, error) {
	repo := NewTaggingRepository(db)
	return repo.FindOrCreate(ctx, taggings)
}

// FindTaggingsByPostID finds taggings by post ID using the repository.
func FindTaggingsByPostID(db *sql.DB, id int) ([]models.Tagging, error) {
	repo := NewTaggingRepository(db)
	return repo.FindByPostID(context.Background(), id)
}

// FindTaggingsByTagID finds taggings by tag ID using the repository.
func FindTaggingsByTagID(ctx context.Context, db *sql.DB, id int) ([]models.Tagging, error) {
	repo := NewTaggingRepository(db)
	return repo.FindByTagID(ctx, id)
}

// DeleteTaggings deletes taggings using the repository.
func DeleteTaggings(ctx context.Context, db *sql.DB, taggings []models.Tagging) error {
	if len(taggings) == 0 {
		return nil
	}

	ids := make([]int, len(taggings))
	for i, tagging := range taggings {
		ids[i] = tagging.ID
	}

	goquDB := goqu.New("postgres", db)
	query := goquDB.From("photos.taggings").Where(goqu.Ex{"id": ids}).Delete()
	_, err := query.Executor().ExecContext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to delete taggings")
	}

	return nil
}

// AllTaggings gets all taggings using the repository.
func AllTaggings(db *sql.DB) ([]models.Tagging, error) {
	repo := NewTaggingRepository(db)
	return repo.All(context.Background())
}
