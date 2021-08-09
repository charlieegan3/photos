package database

import (
	"fmt"

	"github.com/charlieegan3/cms/internal/pkg/models"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type dbDevice struct {
	ID      int    `db:"id"`
	Name    string `db:"name"`
	IconURL string `db:"icon_url"`
}

func newDevice(device dbDevice) models.Device {
	return models.Device{
		ID:      device.ID,
		Name:    device.Name,
		IconURL: device.IconURL,
	}
}

func newDBDevice(device models.Device) dbDevice {
	return dbDevice{
		ID:      device.ID,
		Name:    device.Name,
		IconURL: device.IconURL,
	}
}

func CreateDevices(db *sqlx.DB, devices []models.Device) ([]models.Device, error) {
	dbDevices := []dbDevice{}
	for _, v := range devices {
		dbDevices = append(dbDevices, newDBDevice(v))
	}

	query, args, err := sqlx.Named(
		`INSERT INTO devices (name, icon_url)
		VALUES (:name, :icon_url)
		RETURNING *;`,
		dbDevices)
	if err != nil {
		return []models.Device{}, errors.Wrap(err, "failed to build named sql query")
	}
	fmt.Println(query)
	fmt.Println(args)

	tx := db.MustBegin()
	things, err := tx.QueryRowx(query, args...).SliceScan()
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrap(err, "insert address error")
	}

	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "tx.Commit()")
	}

	returnDevices := []models.Device{}
	for _, v := range things {
		fmt.Println(v)
		// returnDevices = append(returnDevices, newDevice(v))
	}

	return returnDevices, nil
}
