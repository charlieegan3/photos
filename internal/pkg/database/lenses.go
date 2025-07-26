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

func newLens(lens dbLens) models.Lens {
	return models.Lens{
		ID:          lens.ID,
		Name:        lens.Name,
		LensMatches: lens.LensMatches,
		CreatedAt:   lens.CreatedAt,
		UpdatedAt:   lens.UpdatedAt,
	}
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
	if err := insert.ScanStructsContext(ctx, &dbLenses); err != nil {
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
	if err := insert.ScanStructs(&dbLenses); err != nil {
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
	if err := insert.ScanStructs(&dbLenses); err != nil {
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

	if err := selectLenses.ScanStructs(&dbLenses); err != nil {
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
	var dbLenses []dbLens

	goquDB := goqu.New("postgres", db)
	selectLenses := goquDB.From("photos.lenses").
		InnerJoin(goqu.T("medias").Schema("photos"), goqu.On(goqu.Ex{"medias.lens_id": goqu.I("lenses.id")})).
		Select("lenses.*").
		Order(goqu.I("medias.taken_at").Desc()).
		Executor()
	if err := selectLenses.ScanStructs(&dbLenses); err != nil {
		return result, errors.Wrap(err, "failed to select lenses")
	}

	// this is needed in case there are no items added, we don't want to return
	// nil but rather an empty slice
	results := []models.Lens{}
	for _, v := range dbLenses {
		results = append(results, newLens(v))
	}

	if len(results) < 1 {
		return result, nil
	}

	result = results[0]
	return result, nil
}

func DeleteLenses(ctx context.Context, db *sql.DB, lenses []models.Lens) (err error) {
	var ids []int64
	for _, d := range lenses {
		ids = append(ids, d.ID)
	}

	goquDB := goqu.New("postgres", db)
	del, _, err := goquDB.Delete("photos.lenses").Where(
		goqu.Ex{"id": ids},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("failed to build lenses delete query: %s", err)
	}
	_, err = db.ExecContext(ctx, del)
	if err != nil {
		return fmt.Errorf("failed to delete lenses: %s", err)
	}

	return nil
}

// UpdateLenses is not implemented as a single SQL query since update many in
// place is not supported by goqu and it wasn't worth the work (TODO)
func UpdateLenses(ctx context.Context, db *sql.DB, lenses []models.Lens) (results []models.Lens, err error) {
	records := []goqu.Record{}
	for _, v := range lenses {
		d := newDBLens(v)
		records = append(records, d.ToRecord(true))
	}

	goquDB := goqu.New("postgres", db)
	tx, err := goquDB.Begin()
	if err != nil {
		return results, errors.Wrap(err, "failed to open tx for updating lenses")
	}

	for _, record := range records {
		var result dbLens
		update := tx.From("photos.lenses").
			Where(goqu.Ex{"id": record["id"]}).
			Update().
			Set(record).
			Returning(goqu.Star()).
			Executor()
		if _, err = update.ScanStructContext(ctx, &result); err != nil {
			if rErr := tx.Rollback(); rErr != nil {
				return results, errors.Wrap(err, "failed to rollback")
			}
			return results, errors.Wrap(err, "failed to update, rolled back")
		}

		results = append(results, newLens(result))
	}
	if err = tx.Commit(); err != nil {
		return results, errors.Wrap(err, "failed to commit transaction")
	}

	return results, nil
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
	if err := insert.ScanStructsContext(ctx, &dbLenses); err != nil {
		return nil, errors.Wrap(err, "failed to select lenses by lens matches")
	}

	for _, v := range dbLenses {
		l := newLens(v)
		return &l, nil
	}

	return nil, nil
}

func LensPosts(ctx context.Context, db *sql.DB, lensID int64) (results []models.Post, err error) {
	var dbPosts []dbPost

	goquDB := goqu.New("postgres", db)
	selectPosts := goquDB.From("photos.lenses").
		InnerJoin(goqu.T("medias").Schema("photos"), goqu.On(goqu.Ex{"medias.lens_id": goqu.I("lenses.id")})).
		InnerJoin(goqu.T("posts").Schema("photos"), goqu.On(goqu.Ex{"posts.media_id": goqu.I("medias.id")})).
		Select("posts.*").
		Where(goqu.Ex{"lenses.id": lensID}).
		Order(goqu.I("posts.publish_date").Desc()).
		Executor()
	if err := selectPosts.ScanStructsContext(ctx, &dbPosts); err != nil {
		return results, errors.Wrap(err, "failed to select posts")
	}

	// this is needed in case there are no items added, we don't want to return
	// nil but rather an empty slice
	for i := range dbPosts {
		results = append(results, newPost(dbPosts[i]))
	}

	return results, nil
}
