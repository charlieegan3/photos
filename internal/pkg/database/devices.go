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

var ErrDeviceNotFound = errors.New("device not found")

type dbDevice struct {
	ID           int64  `db:"id"`
	Name         string `db:"name"`
	Slug         string `db:"slug"`
	ModelMatches string `db:"model_matches"`
	IconKind     string `db:"icon_kind"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (d *dbDevice) ToRecord(includeID bool) goqu.Record {
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

func (d *dbDevice) ToModel() models.Device {
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
	return (&device).ToModel()
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

func CreateDevices(ctx context.Context, db *sql.DB, devices []models.Device) (results []models.Device, err error) {
	records := []goqu.Record{}
	for _, v := range devices {
		d := newDBDevice(v)
		records = append(records, d.ToRecord(false))
	}

	var dbDevices []dbDevice

	goquDB := goqu.New("postgres", db)
	insert := goquDB.Insert("photos.devices").Returning(goqu.Star()).Rows(records).Executor()
	err = insert.ScanStructsContext(ctx, &dbDevices)
	if err != nil {
		return results, errors.Wrap(err, "failed to insert devices")
	}

	for _, v := range dbDevices {
		results = append(results, newDevice(v))
	}

	return results, nil
}

func FindDevicesByID(ctx context.Context, db *sql.DB, id []int64) (results []models.Device, err error) {
	var dbDevices []dbDevice

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("photos.devices").Select("*").Where(goqu.Ex{"id": id}).Executor()
	err = insert.ScanStructsContext(ctx, &dbDevices)
	if err != nil {
		return results, errors.Wrap(err, "failed to select devices by slug")
	}

	for _, v := range dbDevices {
		results = append(results, newDevice(v))
	}

	return results, nil
}

func FindDevicesByName(ctx context.Context, db *sql.DB, name string) (results []models.Device, err error) {
	var dbDevices []dbDevice

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("photos.devices").Select("*").Where(goqu.Ex{"name": name}).Executor()
	err = insert.ScanStructsContext(ctx, &dbDevices)
	if err != nil {
		return results, errors.Wrap(err, "failed to select devices by slug")
	}

	for _, v := range dbDevices {
		results = append(results, newDevice(v))
	}

	return results, nil
}

func FindDeviceByModelMatches(ctx context.Context, db *sql.DB, modelMatch string) (result *models.Device, err error) {
	var dbDevices []dbDevice

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("photos.devices").
		Select("*").
		Where(goqu.L(`"model_matches" ILIKE ?`, modelMatch)).
		Limit(1).
		Executor()
	err = insert.ScanStructsContext(ctx, &dbDevices)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select devices by slug")
	}

	for _, v := range dbDevices {
		d := newDevice(v)
		return &d, nil
	}

	return nil, ErrDeviceNotFound
}

func AllDevices(ctx context.Context, db *sql.DB) (results []models.Device, err error) {
	var dbDevices []dbDevice

	goquDB := goqu.New("postgres", db)
	selectDevices := goquDB.From("photos.devices").
		FullOuterJoin(goqu.T("medias").Schema("photos"), goqu.On(goqu.Ex{"medias.device_id": goqu.I("devices.id")})).
		FullOuterJoin(goqu.T("posts").Schema("photos"), goqu.On(goqu.Ex{"posts.media_id": goqu.I("medias.id")})).
		Select(
			"devices.*",
		).
		Order(
			goqu.L("MAX(coalesce(posts.publish_date, timestamp with time zone 'epoch'))").Desc(),
		).
		GroupBy(goqu.I("devices.id")).
		Executor()

	err = selectDevices.ScanStructsContext(ctx, &dbDevices)
	if err != nil {
		return results, errors.Wrap(err, "failed to select devices")
	}

	// this is needed in case there are no items added, we don't want to return
	// nil but rather an empty slice
	results = []models.Device{}
	for _, v := range dbDevices {
		results = append(results, newDevice(v))
	}

	return results, nil
}

func MostRecentlyUsedDevice(ctx context.Context, db *sql.DB) (result models.Device, err error) {
	return mostRecentlyUsedEntity[models.Device, dbDevice](ctx, db, "devices", "medias.device_id", newDevice)
}

func DeleteDevices(ctx context.Context, db *sql.DB, devices []models.Device) (err error) {
	ids := make([]int64, len(devices))
	for i, d := range devices {
		ids[i] = d.ID
	}

	goquDB := goqu.New("postgres", db)
	del, _, err := goquDB.Delete("photos.devices").Where(
		goqu.Ex{"id": ids},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("failed to build devices delete query: %w", err)
	}
	_, err = db.ExecContext(ctx, del)
	if err != nil {
		return fmt.Errorf("failed to delete devices: %w", err)
	}

	return nil
}

// UpdateDevices is not implemented as a single SQL query since update many in
// place is not supported by goqu and it wasn't worth the work (TODO).
func UpdateDevices(ctx context.Context, db *sql.DB, devices []models.Device) ([]models.Device, error) {
	return BulkUpdate(ctx, db, "photos.devices", devices, newDBDevice)
}

func DevicePosts(ctx context.Context, db *sql.DB, deviceID int64) (results []models.Post, err error) {
	return entityPosts(ctx, db, "devices", "medias.device_id", "devices.id", deviceID)
}
