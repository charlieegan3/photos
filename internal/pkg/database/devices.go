package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"

	"github.com/charlieegan3/photos/internal/pkg/models"
)

type dbDevice struct {
	ID           int64  `db:"id"`
	Name         string `db:"name"`
	Slug         string `db:"slug"`
	ModelMatches string `db:"model_matches"`
	IconKind     string `db:"icon_kind"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (d dbDevice) ToRecord(includeID bool) goqu.Record {
	record := goqu.Record{
		"name":          d.Name,
		"icon_kind":     d.IconKind,
		"model_matches": d.ModelMatches,
	}

	if includeID {
		record["id"] = d.ID
	}

	return record
}

func (d dbDevice) ToModel() models.Device {
	return models.Device{
		ID:           d.ID,
		Name:         d.Name,
		Slug:         d.Slug,
		ModelMatches: d.ModelMatches,
		IconKind:     d.IconKind,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

func newDevice(device dbDevice) models.Device {
	return device.ToModel()
}

func newDBDevice(device models.Device) dbDevice {
	return dbDevice{
		ID:           device.ID,
		Name:         device.Name,
		ModelMatches: device.ModelMatches,
		IconKind:     device.IconKind,
		CreatedAt:    device.CreatedAt,
		UpdatedAt:    device.UpdatedAt,
	}
}

// DeviceRepository provides device-specific database operations.
type DeviceRepository struct {
	*BaseRepository[models.Device, dbDevice]
}

// NewDeviceRepository creates a new device repository instance.
func NewDeviceRepository(db *sql.DB) *DeviceRepository {
	return &DeviceRepository{
		BaseRepository: NewBaseRepository(db, "devices", newDevice, newDBDevice, "created_at"),
	}
}

// FindByModelMatches finds a device using ILIKE pattern matching on model_matches field.
func (r *DeviceRepository) FindByModelMatches(ctx context.Context, modelMatch string) (*models.Device, error) {
	if modelMatch == "" {
		return nil, errors.New("modelMatch cannot be empty for matching")
	}

	var dbDevices []dbDevice

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select("*").
		Where(goqu.L(`"model_matches" ILIKE ?`, "%"+modelMatch+"%")).
		Limit(1).
		Executor()

	err := query.ScanStructsContext(ctx, &dbDevices)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to select device by model matches for %s", modelMatch)
	}

	if len(dbDevices) == 0 {
		return nil, sql.ErrNoRows
	}

	device := newDevice(dbDevices[0])
	return &device, nil
}

// All retrieves all devices with complex join to posts and medias, ordered by most recent post.
func (r *DeviceRepository) All(ctx context.Context) ([]models.Device, error) {
	return r.AllWithMediaJoins(ctx, "medias.device_id", "devices")
}

// MostRecentlyUsed returns the device that was most recently used in media.
func (r *DeviceRepository) MostRecentlyUsed(ctx context.Context) (models.Device, error) {
	return mostRecentlyUsedEntity[models.Device, dbDevice](ctx, r.db, r.tableName, "medias.device_id", newDevice)
}

// Posts returns all posts associated with a device.
func (r *DeviceRepository) Posts(ctx context.Context, deviceID int64) ([]models.Post, error) {
	return entityPosts(ctx, r.db, r.tableName, "medias.device_id", "devices.id", deviceID)
}

// Legacy function wrappers for backward compatibility with test files.
// These should be removed after all tests are updated.

// CreateDevices creates multiple devices using the repository.
func CreateDevices(ctx context.Context, db *sql.DB, devices []models.Device) ([]models.Device, error) {
	repo := NewDeviceRepository(db)
	return repo.Create(ctx, devices)
}

// FindDevicesByID finds devices by their IDs using the repository.
func FindDevicesByID(ctx context.Context, db *sql.DB, ids []int64) ([]models.Device, error) {
	repo := NewDeviceRepository(db)
	return repo.FindByIDs(ctx, ids)
}

// FindDevicesByName finds devices by name using the repository.
func FindDevicesByName(ctx context.Context, db *sql.DB, name string) ([]models.Device, error) {
	repo := NewDeviceRepository(db)
	return repo.FindByField(ctx, "name", name)
}

// FindDeviceByModelMatches finds a device by model matches using the repository.
func FindDeviceByModelMatches(ctx context.Context, db *sql.DB, modelMatch string) (*models.Device, error) {
	repo := NewDeviceRepository(db)
	return repo.FindByModelMatches(ctx, modelMatch)
}

// AllDevices gets all devices using the repository.
func AllDevices(ctx context.Context, db *sql.DB) ([]models.Device, error) {
	repo := NewDeviceRepository(db)
	return repo.All(ctx)
}

// MostRecentlyUsedDevice gets the most recently used device using the repository.
func MostRecentlyUsedDevice(ctx context.Context, db *sql.DB) (models.Device, error) {
	repo := NewDeviceRepository(db)
	return repo.MostRecentlyUsed(ctx)
}

// DeleteDevices deletes devices using the repository.
func DeleteDevices(ctx context.Context, db *sql.DB, devices []models.Device) error {
	repo := NewDeviceRepository(db)
	return repo.Delete(ctx, devices)
}

// UpdateDevices updates devices using the repository.
func UpdateDevices(ctx context.Context, db *sql.DB, devices []models.Device) ([]models.Device, error) {
	repo := NewDeviceRepository(db)
	return repo.Update(ctx, devices)
}

// DevicePosts returns posts associated with a device using the repository.
func DevicePosts(ctx context.Context, db *sql.DB, deviceID int64) ([]models.Post, error) {
	repo := NewDeviceRepository(db)
	return repo.Posts(ctx, deviceID)
}
