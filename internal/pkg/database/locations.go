package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"

	"github.com/charlieegan3/photos/internal/pkg/models"
)

type dbLocation struct {
	ID        int     `db:"id"`
	Name      string  `db:"name"`
	Slug      string  `db:"slug"`
	Latitude  float64 `db:"latitude"`
	Longitude float64 `db:"longitude"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (d dbLocation) ToRecord(includeID bool) goqu.Record {
	record := goqu.Record{
		"name":      d.Name,
		"latitude":  d.Latitude,
		"longitude": d.Longitude,
	}

	if includeID {
		record["id"] = d.ID
	}

	return record
}

func (d dbLocation) ToModel() models.Location {
	return models.Location{
		ID:        d.ID,
		Name:      d.Name,
		Slug:      d.Slug,
		Latitude:  d.Latitude,
		Longitude: d.Longitude,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func newLocation(location dbLocation) models.Location {
	return location.ToModel()
}

func newDBLocation(location models.Location) dbLocation {
	return dbLocation{
		ID:        location.ID,
		Name:      location.Name,
		Slug:      location.Slug,
		Latitude:  location.Latitude,
		Longitude: location.Longitude,
		CreatedAt: location.CreatedAt,
		UpdatedAt: location.UpdatedAt,
	}
}

// LocationRepository provides location-specific database operations.
type LocationRepository struct {
	*BaseRepository[models.Location, dbLocation]
}

// NewLocationRepository creates a new location repository instance.
func NewLocationRepository(db *sql.DB) *LocationRepository {
	return &LocationRepository{
		BaseRepository: NewBaseRepository(db, "locations", newLocation, newDBLocation, "created_at"),
	}
}

// All retrieves all locations ordered by name (original behavior).
func (r *LocationRepository) All(ctx context.Context) ([]models.Location, error) {
	var dbLocations []dbLocation

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select("*").
		Order(goqu.I("name").Asc()).
		Executor()

	err := query.ScanStructsContext(ctx, &dbLocations)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select all locations")
	}

	results := make([]models.Location, 0, len(dbLocations))
	for _, v := range dbLocations {
		results = append(results, newLocation(v))
	}

	return results, nil
}

// Posts returns all posts associated with a location.
func (r *LocationRepository) Posts(ctx context.Context, locationID int64) ([]models.Post, error) {
	return entityPosts(ctx, r.db, r.tableName, "posts.location_id", "locations.id", locationID)
}

// Legacy function wrappers for backward compatibility with test files.
// These should be removed after all tests are updated.

// CreateLocations creates multiple locations using the repository.
func CreateLocations(ctx context.Context, db *sql.DB, locations []models.Location) ([]models.Location, error) {
	repo := NewLocationRepository(db)
	return repo.Create(ctx, locations)
}

// FindLocationsByID finds locations by their IDs using the repository.
func FindLocationsByID(ctx context.Context, db *sql.DB, ids []int) ([]models.Location, error) {
	if len(ids) == 0 {
		return []models.Location{}, nil
	}

	var dbLocations []dbLocation

	goquDB := goqu.New("postgres", db)
	query := goquDB.From("photos.locations").Select("*").Where(goqu.Ex{"id": ids}).Executor()
	err := query.ScanStructsContext(ctx, &dbLocations)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select locations by id")
	}

	results := make([]models.Location, 0, len(dbLocations))
	for _, location := range dbLocations {
		results = append(results, newLocation(location))
	}

	return results, nil
}

// FindLocationsByName finds locations by name using the repository.
func FindLocationsByName(ctx context.Context, db *sql.DB, name string) ([]models.Location, error) {
	repo := NewLocationRepository(db)
	return repo.FindByField(ctx, "name", name)
}

// AllLocations gets all locations using the repository.
func AllLocations(ctx context.Context, db *sql.DB) ([]models.Location, error) {
	repo := NewLocationRepository(db)
	return repo.All(ctx)
}

// DeleteLocations deletes locations using the repository.
func DeleteLocations(ctx context.Context, db *sql.DB, locations []models.Location) error {
	repo := NewLocationRepository(db)
	return repo.Delete(ctx, locations)
}

// UpdateLocations updates locations using the repository.
func UpdateLocations(ctx context.Context, db *sql.DB, locations []models.Location) ([]models.Location, error) {
	repo := NewLocationRepository(db)
	return repo.Update(ctx, locations)
}

// LocationPosts returns posts associated with a location using the repository.
func LocationPosts(ctx context.Context, db *sql.DB, locationID int64) ([]models.Post, error) {
	repo := NewLocationRepository(db)
	return repo.Posts(ctx, locationID)
}

// MergeLocations merges two locations by moving all posts from the second location to the first.
func MergeLocations(ctx context.Context, db *sql.DB, locationName, oldLocationName string) (int, error) {
	goquDB := goqu.New("postgres", db)

	var id, oldID int

	tx, err := goquDB.Begin()
	if err != nil {
		return 0, errors.Wrap(err, "failed to open transaction")
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Get the target location ID
	newLocationID := tx.From("photos.locations").
		Select("id").
		Where(goqu.Ex{"name": locationName}).Executor()

	result, err := newLocationID.ScanValContext(ctx, &id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil // Return 0 with no error when target location is missing
		}
		return 0, errors.Wrap(err, "failed to get new location ID")
	}
	if !result {
		return 0, nil // Return 0 with no error when target location is missing
	}

	// Get the old location ID
	oldLocationID := tx.From("photos.locations").
		Select("id").
		Where(goqu.Ex{"name": oldLocationName}).Executor()
	result, err = oldLocationID.ScanVal(&oldID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil // Return 0 with no error when old location is missing
		}
		return 0, errors.Wrap(err, "failed to get old location ID")
	}
	if !result {
		return 0, nil // Return 0 with no error when old location is missing
	}

	// Update posts to point to the new location
	updatePosts := tx.Update("photos.posts").
		Where(goqu.Ex{"location_id": oldID}).
		Set(map[string]interface{}{"location_id": id}).
		Executor()
	_, err = updatePosts.Exec()
	if err != nil {
		return 0, errors.Wrap(err, "failed to update posts to merged location")
	}

	// Delete the old location
	deleteLocation := tx.Delete("photos.locations").
		Where(goqu.Ex{"id": oldID}).Executor()
	_, err = deleteLocation.Exec()
	if err != nil {
		return 0, errors.Wrap(err, "failed to delete old location")
	}

	err = tx.Commit()
	if err != nil {
		return 0, errors.Wrap(err, "failed to commit transaction")
	}

	return id, nil
}

// NearbyLocations finds locations near the given coordinates.
func NearbyLocations(db *sql.DB, lat, lon float64) ([]models.Location, error) {
	var dbLocations []dbLocation

	goquDB := goqu.New("postgres", db)
	sub1 := goquDB.From(goqu.T("locations").Schema("photos")).
		Select("*", goqu.L("calculate_distance(?,?, locations.latitude,locations.longitude, 'K')", lat, lon))

	sub2 := goquDB.From(sub1).
		Select("*").
		Where(goqu.I("calculate_distance").Lt(10)).
		Order(goqu.I("calculate_distance").Asc()).
		Limit(10)

	sel := goquDB.From(sub2).
		Select("id",
			"name",
			"slug",
			"latitude",
			"longitude",
			"created_at",
			"updated_at").
		Limit(10)

	err := sel.Executor().ScanStructs(&dbLocations)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select locations")
	}

	// this is needed in case there are no items added, we don't want to return
	// nil but rather an empty slice
	results := make([]models.Location, 0, len(dbLocations))
	for _, v := range dbLocations {
		results = append(results, newLocation(v))
	}

	return results, nil
}
