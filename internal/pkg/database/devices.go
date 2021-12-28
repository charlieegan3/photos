package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"

	"github.com/charlieegan3/photos/cms/internal/pkg/models"
)

type dbDevice struct {
	ID       int    `db:"id"`
	Name     string `db:"name"`
	Slug     string `db:"slug"`
	IconKind string `db:"icon_kind"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (d *dbDevice) ToRecord(includeID bool) goqu.Record {
	record := goqu.Record{
		"name":      d.Name,
		"icon_kind": d.IconKind,
	}

	if includeID {
		record["id"] = d.ID
	}

	return record
}

func newDevice(device dbDevice) models.Device {
	return models.Device{
		ID:        device.ID,
		Name:      device.Name,
		Slug:      device.Slug,
		IconKind:  device.IconKind,
		CreatedAt: device.CreatedAt,
		UpdatedAt: device.UpdatedAt,
	}
}

func newDBDevice(device models.Device) dbDevice {
	return dbDevice{
		ID:        device.ID,
		Name:      device.Name,
		IconKind:  device.IconKind,
		CreatedAt: device.CreatedAt,
		UpdatedAt: device.UpdatedAt,
	}
}

func CreateDevices(db *sql.DB, devices []models.Device) (results []models.Device, err error) {
	records := []goqu.Record{}
	for _, v := range devices {
		d := newDBDevice(v)
		records = append(records, d.ToRecord(false))
	}

	var dbDevices []dbDevice

	goquDB := goqu.New("postgres", db)
	insert := goquDB.Insert("devices").Returning(goqu.Star()).Rows(records).Executor()
	if err := insert.ScanStructs(&dbDevices); err != nil {
		return results, errors.Wrap(err, "failed to insert devices")
	}

	for _, v := range dbDevices {
		results = append(results, newDevice(v))
	}

	return results, nil
}

func FindDevicesByID(db *sql.DB, id int) (results []models.Device, err error) {
	var dbDevices []dbDevice

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("devices").Select("*").Where(goqu.Ex{"id": id}).Executor()
	if err := insert.ScanStructs(&dbDevices); err != nil {
		return results, errors.Wrap(err, "failed to select devices by slug")
	}

	for _, v := range dbDevices {
		results = append(results, newDevice(v))
	}

	return results, nil
}

func FindDevicesByName(db *sql.DB, name string) (results []models.Device, err error) {
	var dbDevices []dbDevice

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("devices").Select("*").Where(goqu.Ex{"name": name}).Executor()
	if err := insert.ScanStructs(&dbDevices); err != nil {
		return results, errors.Wrap(err, "failed to select devices by slug")
	}

	for _, v := range dbDevices {
		results = append(results, newDevice(v))
	}

	return results, nil
}

func AllDevices(db *sql.DB) (results []models.Device, err error) {
	var dbDevices []dbDevice

	goquDB := goqu.New("postgres", db)
	selectDevices := goquDB.From("devices").
		Select("*").
		Order(goqu.I("name").Asc()).
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

func DeleteDevices(db *sql.DB, devices []models.Device) (err error) {
	var ids []int
	for _, d := range devices {
		ids = append(ids, d.ID)
	}

	goquDB := goqu.New("postgres", db)
	del, _, err := goquDB.Delete("devices").Where(
		goqu.Ex{"id": ids},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("failed to build devices delete query: %s", err)
	}
	_, err = db.Exec(del)
	if err != nil {
		return fmt.Errorf("failed to delete devices: %s", err)
	}

	return nil
}

// UpdateDevices is not implemented as a single SQL query since update many in
// place is not supported by goqu and it wasn't worth the work (TODO)
func UpdateDevices(db *sql.DB, devices []models.Device) (results []models.Device, err error) {
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
		update := tx.From("devices").
			Where(goqu.Ex{"id": record["id"]}).
			Update().
			Set(record).
			Returning(goqu.Star()).
			Executor()
		if _, err = update.ScanStruct(&result); err != nil {
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
