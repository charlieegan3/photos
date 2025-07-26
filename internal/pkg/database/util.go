package database

import (
	"context"
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"

	"github.com/charlieegan3/photos/internal/pkg/models"
)

type DBConverter[T any] interface {
	ToModel() T
	ToRecord(bool) goqu.Record
}

func BulkUpdate[T any, D any, DBPointer interface {
	*D
	DBConverter[T]
}](
	ctx context.Context,
	db *sql.DB,
	tableName string,
	items []T,
	toDBFunc func(T) D,
) ([]T, error) {
	records := make([]goqu.Record, 0, len(items))
	for _, item := range items {
		dbItem := toDBFunc(item)
		records = append(records, DBPointer(&dbItem).ToRecord(true))
	}

	goquDB := goqu.New("postgres", db)
	tx, err := goquDB.Begin()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open tx for updating %s", tableName)
	}

	results := make([]T, 0, len(records))
	for _, record := range records {
		var result D
		update := tx.From(tableName).
			Where(goqu.Ex{"id": record["id"]}).
			Update().
			Set(record).
			Returning(goqu.Star()).
			Executor()

		_, err = update.ScanStructContext(ctx, &result)
		if err != nil {
			rErr := tx.Rollback()
			if rErr != nil {
				return nil, errors.Wrap(err, "failed to rollback")
			}
			return nil, errors.Wrap(err, "failed to update, rolled back")
		}

		results = append(results, DBPointer(&result).ToModel())
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	return results, nil
}

func entityPosts(
	ctx context.Context,
	db *sql.DB,
	entityTable, joinColumn, entityIDColumn string,
	entityID int64,
) (results []models.Post, err error) {
	var dbPosts []dbPost

	goquDB := goqu.New("postgres", db)
	selectPosts := goquDB.From(goqu.T(entityTable).Schema("photos")).
		InnerJoin(goqu.T("medias").Schema("photos"), goqu.On(goqu.Ex{joinColumn: goqu.I(entityTable + ".id")})).
		InnerJoin(goqu.T("posts").Schema("photos"), goqu.On(goqu.Ex{"posts.media_id": goqu.I("medias.id")})).
		Select("posts.*").
		Where(goqu.Ex{entityIDColumn: entityID}).
		Order(goqu.I("posts.publish_date").Desc()).
		Executor()
	err = selectPosts.ScanStructsContext(ctx, &dbPosts)
	if err != nil {
		return results, errors.Wrap(err, "failed to select posts")
	}

	for i := range dbPosts {
		results = append(results, newPost(dbPosts[i]))
	}

	return results, nil
}

func mostRecentlyUsedEntity[T any, D any](
	ctx context.Context,
	db *sql.DB,
	tableName, joinColumn string,
	converter func(D) T,
) (result T, err error) {
	var dbEntities []D

	goquDB := goqu.New("postgres", db)
	selectEntities := goquDB.From(goqu.T(tableName).Schema("photos")).
		InnerJoin(goqu.T("medias").Schema("photos"), goqu.On(goqu.Ex{joinColumn: goqu.I(tableName + ".id")})).
		Select(tableName + ".*").
		Order(goqu.I("medias.taken_at").Desc()).
		Executor()
	err = selectEntities.ScanStructsContext(ctx, &dbEntities)
	if err != nil {
		return result, errors.Wrapf(err, "failed to select %s", tableName)
	}

	if len(dbEntities) > 0 {
		result = converter(dbEntities[0])
	}

	return result, nil
}
