package database

import (
	"context"
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/pkg/errors"
)

// ModelConverter is a function that creates a model from a database representation.
type ModelConverter[T any, D DBConverter[T]] func(D) T

// DBModelConverter is a function that creates a database representation from a model.
type DBModelConverter[T any, D DBConverter[T]] func(T) D

// BaseRepository provides generic CRUD operations for any entity.
type BaseRepository[T any, D DBConverter[T]] struct {
	db             *sql.DB
	tableName      string
	schema         string
	toModel        ModelConverter[T, D]
	toDB           DBModelConverter[T, D]
	defaultOrderBy string
}

// NewBaseRepository creates a new generic repository instance.
func NewBaseRepository[T any, D DBConverter[T]](
	db *sql.DB,
	tableName string,
	toModel ModelConverter[T, D],
	toDB DBModelConverter[T, D],
	defaultOrderBy string,
) *BaseRepository[T, D] {
	return &BaseRepository[T, D]{
		db:             db,
		tableName:      tableName,
		schema:         "photos",
		toModel:        toModel,
		toDB:           toDB,
		defaultOrderBy: defaultOrderBy,
	}
}

// Create inserts multiple entities and returns them with generated IDs.
func (r *BaseRepository[T, D]) Create(ctx context.Context, entities []T) ([]T, error) {
	if len(entities) == 0 {
		return []T{}, nil
	}

	records := make([]goqu.Record, 0, len(entities))
	for _, entity := range entities {
		dbEntity := r.toDB(entity)
		records = append(records, dbEntity.ToRecord(false))
	}

	goquDB := goqu.New("postgres", r.db)
	tx, err := goquDB.Begin()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open tx for creating %s entities", r.tableName)
	}

	query := tx.Insert(goqu.T(r.tableName).Schema(r.schema)).
		Returning(goqu.Star()).
		Rows(records).
		Executor()

	var dbEntities []D
	err = query.ScanStructsContext(ctx, &dbEntities)
	if err != nil {
		rErr := tx.Rollback()
		if rErr != nil {
			return nil, errors.Wrapf(err, "failed to rollback %s create transaction: %v", r.tableName, rErr)
		}
		return nil, errors.Wrapf(err, "failed to insert %s entities, rolled back", r.tableName)
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to commit %s create transaction", r.tableName)
	}

	results := make([]T, 0, len(dbEntities))
	for _, dbEntity := range dbEntities {
		results = append(results, r.toModel(dbEntity))
	}

	return results, nil
}

// FindByID retrieves an entity by its ID.
func (r *BaseRepository[T, D]) FindByID(ctx context.Context, id int64) (*T, error) {
	results, err := r.FindByIDs(ctx, []int64{id})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, sql.ErrNoRows
	}
	return &results[0], nil
}

// FindByIDs retrieves multiple entities by their IDs.
func (r *BaseRepository[T, D]) FindByIDs(ctx context.Context, ids []int64) ([]T, error) {
	if len(ids) == 0 {
		return []T{}, nil
	}

	// Filter out non-positive IDs
	validIDs := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id > 0 {
			validIDs = append(validIDs, id)
		}
	}

	if len(validIDs) == 0 {
		return []T{}, nil
	}

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select("*").
		Where(goqu.Ex{"id": validIDs}).
		Executor()

	var dbEntities []D
	err := query.ScanStructsContext(ctx, &dbEntities)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to select %s entities by IDs", r.tableName)
	}

	results := make([]T, 0, len(dbEntities))
	for _, dbEntity := range dbEntities {
		results = append(results, r.toModel(dbEntity))
	}

	return results, nil
}

// FindByField retrieves entities by a specific field value.
func (r *BaseRepository[T, D]) FindByField(ctx context.Context, field string, value interface{}) ([]T, error) {
	if field == "" {
		return nil, errors.Errorf("field name cannot be empty for %s entity search", r.tableName)
	}

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select("*").
		Where(goqu.Ex{field: value})

	if r.defaultOrderBy != "" {
		query = query.Order(goqu.I(r.defaultOrderBy).Desc())
	}

	executor := query.Executor()

	var dbEntities []D
	err := executor.ScanStructsContext(ctx, &dbEntities)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to select %s entities by field %s", r.tableName, field)
	}

	results := make([]T, 0, len(dbEntities))
	for _, dbEntity := range dbEntities {
		results = append(results, r.toModel(dbEntity))
	}

	return results, nil
}

// Exists checks if an entity with the given ID exists.
func (r *BaseRepository[T, D]) Exists(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, errors.Errorf("invalid ID %d for %s entity: ID must be positive", id, r.tableName)
	}

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select(goqu.COUNT("*")).
		Where(goqu.Ex{"id": id}).
		Executor()

	var count int64
	_, err := query.ScanValContext(ctx, &count)
	if err != nil {
		return false, errors.Wrapf(err, "failed to check existence of %s entity with ID %d", r.tableName, id)
	}

	return count > 0, nil
}

// Count returns the total number of entities.
func (r *BaseRepository[T, D]) Count(ctx context.Context) (int64, error) {
	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select(goqu.COUNT("*")).
		Executor()

	var count int64
	_, err := query.ScanValContext(ctx, &count)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to count %s entities", r.tableName)
	}

	return count, nil
}

// All retrieves all entities.
func (r *BaseRepository[T, D]) All(ctx context.Context) ([]T, error) {
	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select("*")

	if r.defaultOrderBy != "" {
		query = query.Order(goqu.I(r.defaultOrderBy).Desc())
	}

	executor := query.Executor()

	var dbEntities []D
	err := executor.ScanStructsContext(ctx, &dbEntities)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to select all %s entities", r.tableName)
	}

	results := make([]T, 0, len(dbEntities)) // Always return empty slice, not nil
	for _, dbEntity := range dbEntities {
		results = append(results, r.toModel(dbEntity))
	}

	return results, nil
}

// Delete removes entities by extracting their IDs.
func (r *BaseRepository[T, D]) Delete(ctx context.Context, entities []T) error {
	if len(entities) == 0 {
		return nil
	}

	// Extract IDs from entities and track invalid ones
	ids := make([]int64, 0, len(entities))
	invalidCount := 0

	for _, entity := range entities {
		dbEntity := r.toDB(entity)
		record := dbEntity.ToRecord(true)
		switch v := record["id"].(type) {
		case int64:
			if v > 0 {
				ids = append(ids, v)
			} else {
				invalidCount++
			}
		case int:
			if v > 0 {
				ids = append(ids, int64(v))
			} else {
				invalidCount++
			}
		default:
			invalidCount++
		}
	}

	if invalidCount > 0 {
		return errors.Errorf("cannot delete %s entities: %d entities have invalid or missing IDs", r.tableName, invalidCount)
	}

	if len(ids) == 0 {
		return nil
	}

	goquDB := goqu.New("postgres", r.db)
	tx, err := goquDB.Begin()
	if err != nil {
		return errors.Wrapf(err, "failed to open tx for deleting %s entities", r.tableName)
	}

	// Use bulk DELETE with IN clause for better performance
	query := tx.From(goqu.T(r.tableName).Schema(r.schema)).
		Where(goqu.Ex{"id": ids}).
		Delete()

	executor := query.Executor()
	_, err = executor.ExecContext(ctx)
	if err != nil {
		rErr := tx.Rollback()
		if rErr != nil {
			return errors.Wrapf(err, "failed to rollback %s delete transaction: %v", r.tableName, rErr)
		}
		return errors.Wrapf(err, "failed to delete %s entities, rolled back", r.tableName)
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrapf(err, "failed to commit %s delete transaction", r.tableName)
	}

	return nil
}

// Update updates multiple entities.
func (r *BaseRepository[T, D]) Update(ctx context.Context, entities []T) ([]T, error) {
	if len(entities) == 0 {
		return []T{}, nil
	}

	goquDB := goqu.New("postgres", r.db)
	tx, err := goquDB.Begin()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open tx for updating %s", r.tableName)
	}

	results := make([]T, 0, len(entities))
	for _, entity := range entities {
		dbEntity := r.toDB(entity)
		record := dbEntity.ToRecord(true)

		update := tx.From(goqu.T(r.tableName).Schema(r.schema)).
			Where(goqu.Ex{"id": record["id"]}).
			Update().
			Set(record).
			Returning(goqu.Star()).
			Executor()

		var result D
		_, err = update.ScanStructContext(ctx, &result)
		if err != nil {
			rErr := tx.Rollback()
			if rErr != nil {
				return nil, errors.Wrapf(err, "failed to rollback %s update transaction: %v", r.tableName, rErr)
			}
			return nil, errors.Wrapf(err, "failed to update %s entity, rolled back", r.tableName)
		}

		results = append(results, r.toModel(result))
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to commit %s update transaction", r.tableName)
	}

	return results, nil
}

// AllWithMediaJoins provides a common pattern for entities that need to join with medias and posts
// for ordering by most recent post activity.
func (r *BaseRepository[T, D]) AllWithMediaJoins(
	ctx context.Context, mediaJoinColumn string, entityTableName string,
) ([]T, error) {
	var dbEntities []D

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		LeftJoin(goqu.T("medias").Schema("photos"), goqu.On(goqu.Ex{mediaJoinColumn: goqu.I(entityTableName + ".id")})).
		LeftJoin(goqu.T("posts").Schema("photos"), goqu.On(goqu.Ex{"posts.media_id": goqu.I("medias.id")})).
		Select(entityTableName + ".*").
		Order(goqu.L("MAX(coalesce(posts.publish_date, timestamp with time zone 'epoch'))").Desc()).
		GroupBy(goqu.I(entityTableName + ".id")).
		Executor()

	err := query.ScanStructsContext(ctx, &dbEntities)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to select all %s", r.tableName)
	}

	results := make([]T, 0, len(dbEntities))
	for i := range dbEntities {
		results = append(results, r.toModel(dbEntities[i]))
	}

	return results, nil
}
