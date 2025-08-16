package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"

	"github.com/charlieegan3/photos/internal/pkg/models"
)

type dbCollection struct {
	ID int `db:"id"`

	Title       string `db:"title"`
	Description string `db:"description"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (d dbCollection) ToRecord(includeID bool) goqu.Record {
	record := goqu.Record{
		"title":       d.Title,
		"description": d.Description,
	}

	if includeID {
		record["id"] = d.ID
	}

	return record
}

func (d dbCollection) ToModel() models.Collection {
	return models.Collection{
		ID: d.ID,

		Title:       d.Title,
		Description: d.Description,

		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func newCollection(collection dbCollection) models.Collection {
	return collection.ToModel()
}

func newDBCollection(collection models.Collection) dbCollection {
	return dbCollection{
		ID:          collection.ID,
		Title:       collection.Title,
		Description: collection.Description,
		CreatedAt:   collection.CreatedAt,
		UpdatedAt:   collection.UpdatedAt,
	}
}

// CollectionRepository provides collection-specific database operations.
type CollectionRepository struct {
	*BaseRepository[models.Collection, dbCollection]
}

// NewCollectionRepository creates a new collection repository instance.
func NewCollectionRepository(db *sql.DB) *CollectionRepository {
	return &CollectionRepository{
		BaseRepository: NewBaseRepository(db, "collections", newCollection, newDBCollection, "created_at"),
	}
}

// AllOrderedByPostCount retrieves all collections ordered by the number of posts in each collection.
func (r *CollectionRepository) AllOrderedByPostCount(ctx context.Context) ([]models.Collection, error) {
	var dbCollections []dbCollection

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		LeftJoin(
			goqu.T("post_collections").Schema(r.schema),
			goqu.On(goqu.Ex{"collections.id": goqu.I("post_collections.collection_id")}),
		).
		Select(
			goqu.I("collections.id"),
			goqu.I("collections.title"),
			goqu.I("collections.description"),
			goqu.I("collections.created_at"),
			goqu.I("collections.updated_at"),
		).
		GroupBy(
			goqu.I("collections.id"),
			goqu.I("collections.title"),
			goqu.I("collections.description"),
			goqu.I("collections.created_at"),
			goqu.I("collections.updated_at"),
		).
		Order(goqu.COUNT("post_collections.id").Desc(), goqu.I("collections.created_at").Desc()).
		Executor()

	err := query.ScanStructsContext(ctx, &dbCollections)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select all collections ordered by post count")
	}

	results := make([]models.Collection, 0, len(dbCollections))
	for i := range dbCollections {
		results = append(results, newCollection(dbCollections[i]))
	}

	return results, nil
}

// Posts returns all posts associated with a collection.
func (r *CollectionRepository) Posts(ctx context.Context, collectionID int) ([]models.Post, error) {
	var dbPosts []dbPost

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T("posts").Schema(r.schema)).
		InnerJoin(
			goqu.T("post_collections").Schema(r.schema),
			goqu.On(goqu.Ex{"posts.id": goqu.I("post_collections.post_id")}),
		).
		Select("posts.*").
		Where(goqu.Ex{"post_collections.collection_id": collectionID}).
		Order(goqu.I("posts.publish_date").Desc()).
		Executor()

	err := query.ScanStructsContext(ctx, &dbPosts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select posts for collection")
	}

	results := make([]models.Post, 0, len(dbPosts))
	for i := range dbPosts {
		results = append(results, newPost(dbPosts[i]))
	}

	return results, nil
}
