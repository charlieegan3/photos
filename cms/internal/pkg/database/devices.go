package database

import (
	"database/sql"
	"time"

	"github.com/charlieegan3/cms/internal/pkg/models"
	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"
)

type dbDevice struct {
	ID      int    `db:"id"`
	Name    string `db:"name"`
	IconURL string `db:"icon_url"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (d *dbDevice) ToRecord(includeID bool) goqu.Record {
	record := goqu.Record{
		"name":     d.Name,
		"icon_url": d.IconURL,
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
		IconURL:   device.IconURL,
		CreatedAt: device.CreatedAt,
		UpdatedAt: device.UpdatedAt,
	}
}

func newDBDevice(device models.Device) dbDevice {
	return dbDevice{
		ID:        device.ID,
		Name:      device.Name,
		IconURL:   device.IconURL,
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

func FindDevicesByName(db *sql.DB, name string) (results []models.Device, err error) {
	var dbDevices []dbDevice

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("devices").Select("*").Where(goqu.Ex{"name": name}).Executor()
	if err := insert.ScanStructs(&dbDevices); err != nil {
		return results, errors.Wrap(err, "failed to select devices by name")
	}

	for _, v := range dbDevices {
		results = append(results, newDevice(v))
	}

	return results, nil
}

func AllDevices(db *sql.DB) (results []models.Device, err error) {
	var dbDevices []dbDevice

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("devices").Select("*").Executor()
	if err := insert.ScanStructs(&dbDevices); err != nil {
		return results, errors.Wrap(err, "failed to select devices")
	}

	for _, v := range dbDevices {
		results = append(results, newDevice(v))
	}

	return results, nil
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
