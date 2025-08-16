package database

import (
	"context"
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"

	"github.com/charlieegan3/photos/internal/pkg/models"
)

type dbPostCollection struct {
	ID int `db:"id"`

	PostID       int `db:"post_id"`
	CollectionID int `db:"collection_id"`
}

func (d dbPostCollection) ToRecord(includeID bool) goqu.Record {
	record := goqu.Record{
		"post_id":       d.PostID,
		"collection_id": d.CollectionID,
	}

	if includeID {
		record["id"] = d.ID
	}

	return record
}

func (d dbPostCollection) ToModel() models.PostCollection {
	return models.PostCollection{
		ID: d.ID,

		PostID:       d.PostID,
		CollectionID: d.CollectionID,
	}
}

func newPostCollection(postCollection dbPostCollection) models.PostCollection {
	return postCollection.ToModel()
}

func newDBPostCollection(postCollection models.PostCollection) dbPostCollection {
	return dbPostCollection{
		ID: postCollection.ID,

		PostID:       postCollection.PostID,
		CollectionID: postCollection.CollectionID,
	}
}

// PostCollectionRepository provides post_collection-specific database operations.
type PostCollectionRepository struct {
	*BaseRepository[models.PostCollection, dbPostCollection]
}

// NewPostCollectionRepository creates a new post collection repository instance.
func NewPostCollectionRepository(db *sql.DB) *PostCollectionRepository {
	return &PostCollectionRepository{
		BaseRepository: NewBaseRepository(db, "post_collections", newPostCollection, newDBPostCollection, "post_id"),
	}
}

// All retrieves all post_collections ordered by post_id ASC for predictable ordering.
func (r *PostCollectionRepository) All(ctx context.Context) ([]models.PostCollection, error) {
	var dbPostCollections []dbPostCollection

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select("*").
		Order(goqu.I("post_id").Asc()).
		Executor()

	err := query.ScanStructsContext(ctx, &dbPostCollections)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select all post_collections")
	}

	results := make([]models.PostCollection, 0, len(dbPostCollections))
	for _, dbPostCollection := range dbPostCollections {
		results = append(results, newPostCollection(dbPostCollection))
	}

	return results, nil
}

// FindByPostID finds post_collections by post ID.
func (r *PostCollectionRepository) FindByPostID(ctx context.Context, id int) ([]models.PostCollection, error) {
	var dbPostCollections []dbPostCollection

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select("*").
		Where(goqu.Ex{"post_id": id}).
		Executor()

	err := query.ScanStructsContext(ctx, &dbPostCollections)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select post_collections by post_id")
	}

	results := make([]models.PostCollection, 0, len(dbPostCollections))
	for _, postCollection := range dbPostCollections {
		results = append(results, newPostCollection(postCollection))
	}

	return results, nil
}

// FindByCollectionID finds post_collections by collection ID.
func (r *PostCollectionRepository) FindByCollectionID(ctx context.Context, id int) ([]models.PostCollection, error) {
	var dbPostCollections []dbPostCollection

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select("*").
		Where(goqu.Ex{"collection_id": id}).
		Executor()

	err := query.ScanStructsContext(ctx, &dbPostCollections)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select post_collections by collection_id")
	}

	results := make([]models.PostCollection, 0, len(dbPostCollections))
	for _, postCollection := range dbPostCollections {
		results = append(results, newPostCollection(postCollection))
	}

	return results, nil
}

// DeleteByPostID deletes all post_collections for a given post ID.
func (r *PostCollectionRepository) DeleteByPostID(ctx context.Context, postID int) error {
	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Where(goqu.Ex{"post_id": postID}).
		Delete()

	_, err := query.Executor().ExecContext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to delete post_collections by post_id")
	}

	return nil
}
