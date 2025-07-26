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

var ErrLensNotFound = errors.New("lens not found")

type dbLens struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`

	LensMatches string `db:"lens_matches"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (d *dbLens) ToRecord(includeID bool) goqu.Record {
	record := goqu.Record{
		"name":         d.Name,
		"lens_matches": d.LensMatches,
	}

	if includeID {
		record["id"] = d.ID
	}

	return record
}

func (d *dbLens) ToModel() models.Lens {
	return models.Lens{
		ID:          d.ID,
		Name:        d.Name,
		LensMatches: d.LensMatches,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func newLens(lens dbLens) models.Lens {
	return (&lens).ToModel()
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

func CreateLenses(ctx context.Context, db *sql.DB, lenses []models.Lens) (results []models.Lens, err error) {
	records := []goqu.Record{}
	for _, v := range lenses {
		d := newDBLens(v)
		records = append(records, d.ToRecord(false))
	}

	var dbLenses []dbLens

	goquDB := goqu.New("postgres", db)
	insert := goquDB.Insert("photos.lenses").Returning(goqu.Star()).Rows(records).Executor()
	err = insert.ScanStructsContext(ctx, &dbLenses)
	if err != nil {
		return results, errors.Wrap(err, "failed to insert lenses")
	}

	for _, v := range dbLenses {
		results = append(results, newLens(v))
	}

	return results, nil
}

func FindLensesByID(ctx context.Context, db *sql.DB, id []int64) (results []models.Lens, err error) {
	var dbLenses []dbLens

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("photos.lenses").Select("*").Where(goqu.Ex{"id": id}).Executor()
	err = insert.ScanStructsContext(ctx, &dbLenses)
	if err != nil {
		return results, errors.Wrap(err, "failed to select lenses by slug")
	}

	for _, v := range dbLenses {
		results = append(results, newLens(v))
	}

	return results, nil
}

func FindLensesByName(ctx context.Context, db *sql.DB, name string) (results []models.Lens, err error) {
	var dbLenses []dbLens

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("photos.lenses").Select("*").Where(goqu.Ex{"name": name}).Executor()
	err = insert.ScanStructsContext(ctx, &dbLenses)
	if err != nil {
		return results, errors.Wrap(err, "failed to select lenses by slug")
	}

	for _, v := range dbLenses {
		results = append(results, newLens(v))
	}

	return results, nil
}

func AllLenses(ctx context.Context, db *sql.DB) (results []models.Lens, err error) {
	var dbLenses []dbLens

	goquDB := goqu.New("postgres", db)
	selectLenses := goquDB.From("photos.lenses").
		FullOuterJoin(goqu.T("medias").Schema("photos"), goqu.On(goqu.Ex{"medias.lens_id": goqu.I("lenses.id")})).
		FullOuterJoin(goqu.T("posts").Schema("photos"), goqu.On(goqu.Ex{"posts.media_id": goqu.I("medias.id")})).
		Select(
			"lenses.*",
		).
		Where(goqu.L("lenses.id IS NOT NULL")).
		Order(
			goqu.L("MAX(coalesce(posts.publish_date, timestamp with time zone 'epoch'))").Desc(),
		).
		GroupBy(goqu.I("lenses.id")).
		Executor()

	err = selectLenses.ScanStructsContext(ctx, &dbLenses)
	if err != nil {
		return results, errors.Wrap(err, "failed to select lenses")
	}

	// this is needed in case there are no items added, we don't want to return
	// nil but rather an empty slice
	results = []models.Lens{}
	for _, v := range dbLenses {
		results = append(results, newLens(v))
	}

	return results, nil
}

func MostRecentlyUsedLens(ctx context.Context, db *sql.DB) (result models.Lens, err error) {
	return mostRecentlyUsedEntity[models.Lens, dbLens](ctx, db, "lenses", "medias.lens_id", newLens)
}

func DeleteLenses(ctx context.Context, db *sql.DB, lenses []models.Lens) (err error) {
	ids := make([]int64, len(lenses))
	for i, d := range lenses {
		ids[i] = d.ID
	}

	goquDB := goqu.New("postgres", db)
	del, _, err := goquDB.Delete("photos.lenses").Where(
		goqu.Ex{"id": ids},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("failed to build lenses delete query: %w", err)
	}
	_, err = db.ExecContext(ctx, del)
	if err != nil {
		return fmt.Errorf("failed to delete lenses: %w", err)
	}

	return nil
}

// UpdateLenses is not implemented as a single SQL query since update many in
// place is not supported by goqu and it wasn't worth the work (TODO).
func UpdateLenses(ctx context.Context, db *sql.DB, lenses []models.Lens) ([]models.Lens, error) {
	return BulkUpdate(ctx, db, "photos.lenses", lenses, newDBLens)
}

func FindLensByLensMatches(ctx context.Context, db *sql.DB, lensMatch string) (result *models.Lens, err error) {
	var dbLenses []dbLens

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("photos.lenses").
		Select("*").
		Where(goqu.L(
			`"lens_matches" ILIKE ?`,
			fmt.Sprintf("%%%s%%", lensMatch),
		)).
		Limit(1).
		Executor()
	err = insert.ScanStructsContext(ctx, &dbLenses)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select lenses by lens matches")
	}

	for _, v := range dbLenses {
		l := newLens(v)
		return &l, nil
	}

	return nil, ErrLensNotFound
}

func LensPosts(ctx context.Context, db *sql.DB, lensID int64) (results []models.Post, err error) {
	return entityPosts(ctx, db, "lenses", "medias.lens_id", "lenses.id", lensID)
}
