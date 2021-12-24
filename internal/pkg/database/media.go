package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"

	"github.com/charlieegan3/photos/cms/internal/pkg/models"
)

type dbMedia struct {
	ID int `db:"id"`

	Kind string `db:"kind"`

	Make  string `db:"make"`
	Model string `db:"model"`

	TakenAt time.Time `db:"taken_at"`

	FNumber      float64 `db:"f_number"`
	ShutterSpeed float64 `db:"shutter_speed"`
	ISOSpeed     int     `db:"iso_speed"`

	Latitude  float64 `db:"latitude"`
	Longitude float64 `db:"longitude"`
	Altitude  float64 `db:"altitude"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	InstagramCode string `db:"instagram_code"`

	DeviceID int `db:"device_id"`
}

func (d *dbMedia) ToRecord(includeID bool) goqu.Record {
	record := goqu.Record{
		"kind":           d.Kind,
		"make":           d.Make,
		"model":          d.Model,
		"taken_at":       d.TakenAt.Format("2006-01-02 15:04:05"), // strip the zone since it's not in exif
		"f_number":       d.FNumber,
		"shutter_speed":  d.ShutterSpeed,
		"iso_speed":      d.ISOSpeed,
		"latitude":       d.Latitude,
		"longitude":      d.Longitude,
		"altitude":       d.Altitude,
		"device_id":      d.DeviceID,
		"instagram_code": d.InstagramCode,
	}

	if includeID {
		record["id"] = d.ID
	}

	return record
}

func newMedia(media dbMedia) models.Media {
	return models.Media{
		ID: media.ID,

		Kind: media.Kind,

		Make:  media.Make,
		Model: media.Model,
		// present as UTC since zone information is missing in EXIF
		TakenAt:      media.TakenAt.UTC(),
		FNumber:      media.FNumber,
		ShutterSpeed: media.ShutterSpeed,
		ISOSpeed:     media.ISOSpeed,
		Latitude:     media.Latitude,
		Longitude:    media.Longitude,
		Altitude:     media.Altitude,

		CreatedAt: media.CreatedAt,
		UpdatedAt: media.UpdatedAt,

		DeviceID: media.DeviceID,

		InstagramCode: media.InstagramCode,
	}
}

func newDBMedia(media models.Media) dbMedia {
	return dbMedia{
		ID: media.ID,

		Kind:         media.Kind,
		Make:         media.Make,
		Model:        media.Model,
		TakenAt:      media.TakenAt,
		FNumber:      media.FNumber,
		ShutterSpeed: media.ShutterSpeed,
		ISOSpeed:     media.ISOSpeed,
		Latitude:     media.Latitude,
		Longitude:    media.Longitude,
		Altitude:     media.Altitude,

		CreatedAt: media.CreatedAt,
		UpdatedAt: media.UpdatedAt,

		DeviceID: media.DeviceID,

		InstagramCode: media.InstagramCode,
	}
}

func CreateMedias(db *sql.DB, medias []models.Media) (results []models.Media, err error) {
	records := []goqu.Record{}
	for _, v := range medias {
		d := newDBMedia(v)
		records = append(records, d.ToRecord(false))
	}

	var dbMedias []dbMedia

	goquDB := goqu.New("postgres", db)
	insert := goquDB.Insert("medias").Returning(goqu.Star()).Rows(records).Executor()
	if err := insert.ScanStructs(&dbMedias); err != nil {
		return results, errors.Wrap(err, "failed to insert medias")
	}

	for _, v := range dbMedias {
		results = append(results, newMedia(v))
	}

	return results, nil
}

func FindMediasByID(db *sql.DB, id int) (results []models.Media, err error) {
	var dbMedias []dbMedia

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("medias").Select("*").Where(goqu.Ex{"id": id}).Executor()
	if err := insert.ScanStructs(&dbMedias); err != nil {
		return results, errors.Wrap(err, "failed to select medias by id")
	}

	for _, v := range dbMedias {
		results = append(results, newMedia(v))
	}

	return results, nil
}

func FindMediasByInstagramCode(db *sql.DB, code string) (results []models.Media, err error) {
	var dbMedias []dbMedia

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("medias").Select("*").Where(goqu.Ex{"instagram_code": code}).Executor()
	if err := insert.ScanStructs(&dbMedias); err != nil {
		return results, errors.Wrap(err, "failed to select medias by id")
	}

	for _, v := range dbMedias {
		results = append(results, newMedia(v))
	}

	return results, nil
}

func AllMedias(db *sql.DB, descending bool) (results []models.Media, err error) {
	var dbMedias []dbMedia

	goquDB := goqu.New("postgres", db)
	insert := goquDB.From("medias").Select("*")

	if descending {
		insert = insert.Order(goqu.I("taken_at").Desc(), goqu.I("created_at").Desc())
	}

	if err := insert.Executor().ScanStructs(&dbMedias); err != nil {
		return results, errors.Wrap(err, "failed to select medias")
	}

	// this is needed in case there are no items added, we don't want to return
	// nil but rather an empty slice
	results = []models.Media{}
	for _, v := range dbMedias {
		results = append(results, newMedia(v))
	}

	return results, nil
}

func DeleteMedias(db *sql.DB, medias []models.Media) (err error) {
	var ids []int
	for _, d := range medias {
		ids = append(ids, d.ID)
	}

	goquDB := goqu.New("postgres", db)
	del, _, err := goquDB.Delete("medias").Where(
		goqu.Ex{"id": ids},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("failed to build medias delete query: %s", err)
	}
	_, err = db.Exec(del)
	if err != nil {
		return fmt.Errorf("failed to delete medias: %s", err)
	}

	return nil
}

// UpdateMedias is not implemented as a single SQL query since update many in
// place is not supported by goqu and it wasn't worth the work (TODO)
func UpdateMedias(db *sql.DB, medias []models.Media) (results []models.Media, err error) {
	records := []goqu.Record{}
	for _, v := range medias {
		d := newDBMedia(v)
		records = append(records, d.ToRecord(true))
	}

	goquDB := goqu.New("postgres", db)
	tx, err := goquDB.Begin()
	if err != nil {
		return results, errors.Wrap(err, "failed to open tx for updating medias")
	}

	for _, record := range records {
		var result dbMedia
		update := tx.From("medias").
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

		results = append(results, newMedia(result))
	}
	if err = tx.Commit(); err != nil {
		return results, errors.Wrap(err, "failed to commit transaction")
	}

	return results, nil
}