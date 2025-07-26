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

func newDevice(device dbDevice) models.Device {
	return models.Device{
		ID:           device.ID,
		Name:         device.Name,
		Slug:         device.Slug,
		ModelMatches: device.ModelMatches,
		IconKind:     device.IconKind,
		CreatedAt:    device.CreatedAt,
		UpdatedAt:    device.UpdatedAt,
	}
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
	if err := insert.ScanStructsContext(ctx, &dbDevices); err != nil {
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
	if err := insert.ScanStructsContext(ctx, &dbDevices); err != nil {
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
	if err := insert.ScanStructsContext(ctx, &dbDevices); err != nil {
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
	if err := insert.ScanStructsContext(ctx, &dbDevices); err != nil {
		return nil, errors.Wrap(err, "failed to select devices by slug")
	}

	for _, v := range dbDevices {
		d := newDevice(v)
		return &d, nil
	}

	return nil, nil
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

	if err := selectDevices.ScanStructs(&dbDevices); err != nil {
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
	var dbDevices []dbDevice

	goquDB := goqu.New("postgres", db)
	selectDevices := goquDB.From("photos.devices").
		InnerJoin(goqu.T("medias").Schema("photos"), goqu.On(goqu.Ex{"medias.device_id": goqu.I("devices.id")})).
		Select("devices.*").
		Order(goqu.I("medias.taken_at").Desc()).
		Executor()
	if err := selectDevices.ScanStructsContext(ctx, &dbDevices); err != nil {
		return result, errors.Wrap(err, "failed to select devices")
	}

	// this is needed in case there are no items added, we don't want to return
	// nil but rather an empty slice
	results := []models.Device{}
	for _, v := range dbDevices {
		results = append(results, newDevice(v))
	}

	if len(results) < 1 {
		return result, nil
	}

	result = results[0]
	return result, nil
}

func DeleteDevices(ctx context.Context, db *sql.DB, devices []models.Device) (err error) {
	var ids []int64
	for _, d := range devices {
		ids = append(ids, d.ID)
	}

	goquDB := goqu.New("postgres", db)
	del, _, err := goquDB.Delete("photos.devices").Where(
		goqu.Ex{"id": ids},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("failed to build devices delete query: %s", err)
	}
	_, err = db.ExecContext(ctx, del)
	if err != nil {
		return fmt.Errorf("failed to delete devices: %s", err)
	}

	return nil
}

// UpdateDevices is not implemented as a single SQL query since update many in
// place is not supported by goqu and it wasn't worth the work (TODO)
func UpdateDevices(ctx context.Context, db *sql.DB, devices []models.Device) (results []models.Device, err error) {
	records := []goqu.Record{}
	for _, v := range devices {
		d := newDBDevice(v)
		records = append(records, d.ToRecord(true))
	}

	goquDB := goqu.New("postgres", db)
	tx, err := goquDB.Begin()
	if err != nil {
		return results, errors.Wrap(err, "failed to open tx for updating devices")
	}

	for _, record := range records {
		var result dbDevice
		update := tx.From("photos.devices").
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

		results = append(results, newDevice(result))
	}
	if err = tx.Commit(); err != nil {
		return results, errors.Wrap(err, "failed to commit transaction")
	}

	return results, nil
}

func DevicePosts(ctx context.Context, db *sql.DB, deviceID int64) (results []models.Post, err error) {
	var dbPosts []dbPost

	goquDB := goqu.New("postgres", db)
	selectPosts := goquDB.From("photos.devices").
		InnerJoin(goqu.T("medias").Schema("photos"), goqu.On(goqu.Ex{"medias.device_id": goqu.I("devices.id")})).
		InnerJoin(goqu.T("posts").Schema("photos"), goqu.On(goqu.Ex{"posts.media_id": goqu.I("medias.id")})).
		Select("posts.*").
		Where(goqu.Ex{"devices.id": deviceID}).
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
