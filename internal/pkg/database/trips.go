package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"

	"github.com/charlieegan3/photos/internal/pkg/models"
)

type dbTrip struct {
	ID int `db:"id"`

	Title       string `db:"title"`
	Description string `db:"description"`

	StartDate time.Time `db:"start_date"`
	EndDate   time.Time `db:"end_date"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (d dbTrip) ToRecord(includeID bool) goqu.Record {
	record := goqu.Record{
		"title":       d.Title,
		"description": d.Description,
		"start_date":  d.StartDate.UTC(),
		"end_date":    d.EndDate.UTC(),
	}

	if includeID {
		record["id"] = d.ID
	}

	return record
}

func (d dbTrip) ToModel() models.Trip {
	return models.Trip{
		ID: d.ID,

		Title:       d.Title,
		Description: d.Description,

		StartDate: d.StartDate.UTC(),
		EndDate:   d.EndDate.UTC(),

		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func newTrip(trip dbTrip) models.Trip {
	return trip.ToModel()
}

func newDBTrip(trip models.Trip) dbTrip {
	return dbTrip{
		ID:          trip.ID,
		Title:       trip.Title,
		Description: trip.Description,
		StartDate:   trip.StartDate.UTC(),
		EndDate:     trip.EndDate.UTC(),
		CreatedAt:   trip.CreatedAt,
		UpdatedAt:   trip.UpdatedAt,
	}
}

// TripRepository provides trip-specific database operations.
type TripRepository struct {
	*BaseRepository[models.Trip, dbTrip]
}

// NewTripRepository creates a new trip repository instance.
func NewTripRepository(db *sql.DB) *TripRepository {
	return &TripRepository{
		BaseRepository: NewBaseRepository(db, "trips", newTrip, newDBTrip, "start_date"),
	}
}

// All retrieves all trips ordered by start date (original behavior).
func (r *TripRepository) All(ctx context.Context) ([]models.Trip, error) {
	var dbTrips []dbTrip

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select("*").
		Order(goqu.I("start_date").Desc()).
		Executor()

	err := query.ScanStructsContext(ctx, &dbTrips)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select all trips")
	}

	results := make([]models.Trip, 0, len(dbTrips))
	for i := range dbTrips {
		results = append(results, newTrip(dbTrips[i]))
	}

	return results, nil
}

// Posts returns all posts associated with a trip.
// NOTE: This relationship may not exist in the current schema.
func (*TripRepository) Posts(_ context.Context, _ int64) ([]models.Post, error) {
	// TODO: Implement trip-post relationship when schema supports it
	return []models.Post{}, nil
}

// Legacy function wrappers for backward compatibility with test files.
// These should be removed after all tests are updated.

// CreateTrips creates multiple trips using the repository.
func CreateTrips(ctx context.Context, db *sql.DB, trips []models.Trip) ([]models.Trip, error) {
	repo := NewTripRepository(db)
	return repo.Create(ctx, trips)
}

// FindTripsByID finds trips by their IDs using the repository.
func FindTripsByID(ctx context.Context, db *sql.DB, ids []int) ([]models.Trip, error) {
	if len(ids) == 0 {
		return []models.Trip{}, nil
	}

	var dbTrips []dbTrip

	goquDB := goqu.New("postgres", db)
	query := goquDB.From("photos.trips").Select("*").Where(goqu.Ex{"id": ids}).Executor()
	err := query.ScanStructsContext(ctx, &dbTrips)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select trips by id")
	}

	results := make([]models.Trip, 0, len(dbTrips))
	for i := range dbTrips {
		results = append(results, newTrip(dbTrips[i]))
	}

	return results, nil
}

// FindTripsByName finds trips by title using the repository.
func FindTripsByName(ctx context.Context, db *sql.DB, title string) ([]models.Trip, error) {
	repo := NewTripRepository(db)
	return repo.FindByField(ctx, "title", title)
}

// AllTrips gets all trips using the repository.
func AllTrips(ctx context.Context, db *sql.DB) ([]models.Trip, error) {
	repo := NewTripRepository(db)
	return repo.All(ctx)
}

// DeleteTrips deletes trips using the repository.
func DeleteTrips(ctx context.Context, db *sql.DB, trips []models.Trip) error {
	repo := NewTripRepository(db)
	return repo.Delete(ctx, trips)
}

// UpdateTrips updates trips using the repository.
func UpdateTrips(ctx context.Context, db *sql.DB, trips []models.Trip) ([]models.Trip, error) {
	repo := NewTripRepository(db)
	return repo.Update(ctx, trips)
}

// TripPosts returns posts associated with a trip using the repository.
func TripPosts(ctx context.Context, db *sql.DB, tripID int64) ([]models.Post, error) {
	repo := NewTripRepository(db)
	return repo.Posts(ctx, tripID)
}
