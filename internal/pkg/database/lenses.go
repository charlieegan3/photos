package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"

	"github.com/charlieegan3/photos/internal/pkg/models"
)

type dbLens struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`

	LensMatches string `db:"lens_matches"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (d dbLens) ToRecord(includeID bool) goqu.Record {
	record := goqu.Record{
		"name":         d.Name,
		"lens_matches": d.LensMatches,
	}

	if includeID {
		record["id"] = d.ID
	}

	return record
}

func (d dbLens) ToModel() models.Lens {
	return models.Lens{
		ID:          d.ID,
		Name:        d.Name,
		LensMatches: d.LensMatches,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func newLens(lens dbLens) models.Lens {
	return lens.ToModel()
}

func newDBLens(lens models.Lens) dbLens {
	return dbLens{
		ID:          lens.ID,
		Name:        lens.Name,
		LensMatches: lens.LensMatches,
		CreatedAt:   lens.CreatedAt,
		UpdatedAt:   lens.UpdatedAt,
	}
}

// LensRepository provides lens-specific database operations.
type LensRepository struct {
	*BaseRepository[models.Lens, dbLens]
}

// NewLensRepository creates a new lens repository instance.
func NewLensRepository(db *sql.DB) *LensRepository {
	return &LensRepository{
		BaseRepository: NewBaseRepository(db, "lenses", newLens, newDBLens, "created_at"),
	}
}

// FindByLensMatches finds a lens using ILIKE pattern matching on lens_matches field.
func (r *LensRepository) FindByLensMatches(ctx context.Context, lensMatch string) (*models.Lens, error) {
	if lensMatch == "" {
		return nil, errors.New("lensMatch cannot be empty for matching")
	}

	var dbLenses []dbLens

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select("*").
		Where(goqu.L(`"lens_matches" ILIKE ?`, "%"+lensMatch+"%")).
		Limit(1).
		Executor()

	err := query.ScanStructsContext(ctx, &dbLenses)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to select lens by lens matches for %s", lensMatch)
	}

	if len(dbLenses) == 0 {
		return nil, sql.ErrNoRows
	}

	lens := newLens(dbLenses[0])
	return &lens, nil
}

// All retrieves all lenses with complex join to posts and medias, ordered by most recent post.
func (r *LensRepository) All(ctx context.Context) ([]models.Lens, error) {
	return r.AllWithMediaJoins(ctx, "medias.lens_id", "lenses")
}

// MostRecentlyUsed returns the lens that was most recently used in media.
func (r *LensRepository) MostRecentlyUsed(ctx context.Context) (models.Lens, error) {
	return mostRecentlyUsedEntity[models.Lens, dbLens](ctx, r.db, r.tableName, "medias.lens_id", newLens)
}

// Posts returns all posts associated with a lens.
func (r *LensRepository) Posts(ctx context.Context, lensID int64) ([]models.Post, error) {
	return entityPosts(ctx, r.db, r.tableName, "medias.lens_id", "lenses.id", lensID)
}

// Legacy function wrappers for backward compatibility with test files.
// These should be removed after all tests are updated.

// CreateLenses creates multiple lenses using the repository.
func CreateLenses(ctx context.Context, db *sql.DB, lenses []models.Lens) ([]models.Lens, error) {
	repo := NewLensRepository(db)
	return repo.Create(ctx, lenses)
}

// FindLensesByID finds lenses by their IDs using the repository.
func FindLensesByID(ctx context.Context, db *sql.DB, ids []int64) ([]models.Lens, error) {
	repo := NewLensRepository(db)
	return repo.FindByIDs(ctx, ids)
}

// FindLensesByName finds lenses by name using the repository.
func FindLensesByName(ctx context.Context, db *sql.DB, name string) ([]models.Lens, error) {
	repo := NewLensRepository(db)
	return repo.FindByField(ctx, "name", name)
}

// FindLensByLensMatches finds a lens by lens matches using the repository.
func FindLensByLensMatches(ctx context.Context, db *sql.DB, lensMatch string) (*models.Lens, error) {
	repo := NewLensRepository(db)
	return repo.FindByLensMatches(ctx, lensMatch)
}

// AllLenses gets all lenses using the repository.
func AllLenses(ctx context.Context, db *sql.DB) ([]models.Lens, error) {
	repo := NewLensRepository(db)
	return repo.All(ctx)
}

// MostRecentlyUsedLens gets the most recently used lens using the repository.
func MostRecentlyUsedLens(ctx context.Context, db *sql.DB) (models.Lens, error) {
	repo := NewLensRepository(db)
	return repo.MostRecentlyUsed(ctx)
}

// DeleteLenses deletes lenses using the repository.
func DeleteLenses(ctx context.Context, db *sql.DB, lenses []models.Lens) error {
	repo := NewLensRepository(db)
	return repo.Delete(ctx, lenses)
}

// UpdateLenses updates lenses using the repository.
func UpdateLenses(ctx context.Context, db *sql.DB, lenses []models.Lens) ([]models.Lens, error) {
	repo := NewLensRepository(db)
	return repo.Update(ctx, lenses)
}

// LensPosts returns posts associated with a lens using the repository.
func LensPosts(ctx context.Context, db *sql.DB, lensID int64) ([]models.Post, error) {
	repo := NewLensRepository(db)
	return repo.Posts(ctx, lensID)
}
