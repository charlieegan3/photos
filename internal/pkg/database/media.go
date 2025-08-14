package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"

	"github.com/charlieegan3/photos/internal/pkg/models"
)

type dbMedia struct {
	ID int `db:"id"`

	Kind string `db:"kind"`

	Make  string `db:"make"`
	Model string `db:"model"`

	Lens        string `db:"lens"`
	FocalLength string `db:"focal_length"`

	TakenAt time.Time `db:"taken_at"`

	FNumber                 float64 `db:"f_number"`
	ExposureTimeNumerator   uint32  `db:"exposure_time_numerator"`
	ExposureTimeDenominator uint32  `db:"exposure_time_denominator"`
	ISOSpeed                int     `db:"iso_speed"`

	Latitude  float64 `db:"latitude"`
	Longitude float64 `db:"longitude"`
	Altitude  float64 `db:"altitude"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	InstagramCode string `db:"instagram_code"`

	DeviceID int64 `db:"device_id"`

	LensID sql.NullInt64 `db:"lens_id"`

	UTCCorrect bool `db:"utc_correct"`

	Width  int `db:"width"`
	Height int `db:"height"`

	Orientation int `db:"orientation"`

	DisplayOffset int `db:"display_offset"`
}

func (d dbMedia) ToRecord(includeID bool) goqu.Record {
	record := goqu.Record{
		"kind":                      d.Kind,
		"make":                      d.Make,
		"model":                     d.Model,
		"lens":                      d.Lens,
		"focal_length":              d.FocalLength,
		"taken_at":                  d.TakenAt.Format("2006-01-02 15:04:05"), // strip the zone since it's not in exif
		"f_number":                  d.FNumber,
		"exposure_time_numerator":   d.ExposureTimeNumerator,
		"exposure_time_denominator": d.ExposureTimeDenominator,
		"iso_speed":                 d.ISOSpeed,
		"latitude":                  d.Latitude,
		"longitude":                 d.Longitude,
		"altitude":                  d.Altitude,
		"device_id":                 d.DeviceID,
		"instagram_code":            d.InstagramCode,
		"utc_correct":               d.UTCCorrect,
		"width":                     d.Width,
		"height":                    d.Height,
		"orientation":               d.Orientation,
		"display_offset":            d.DisplayOffset,
	}

	record["lens_id"] = nil
	if d.LensID.Valid {
		record["lens_id"] = d.LensID.Int64
	}

	if includeID {
		record["id"] = d.ID
	}

	return record
}

func (d dbMedia) ToModel() models.Media {
	media := models.Media{
		ID: d.ID,

		Kind: d.Kind,

		Make:  d.Make,
		Model: d.Model,

		Lens:        d.Lens,
		FocalLength: d.FocalLength,

		// present as UTC since zone information is missing in EXIF
		TakenAt:                 d.TakenAt.UTC(),
		FNumber:                 d.FNumber,
		ExposureTimeNumerator:   d.ExposureTimeNumerator,
		ExposureTimeDenominator: d.ExposureTimeDenominator,
		ISOSpeed:                d.ISOSpeed,
		Latitude:                d.Latitude,
		Longitude:               d.Longitude,
		Altitude:                d.Altitude,

		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,

		DeviceID: d.DeviceID,

		InstagramCode: d.InstagramCode,

		UTCCorrect: d.UTCCorrect,

		Width:  d.Width,
		Height: d.Height,

		Orientation: d.Orientation,

		DisplayOffset: d.DisplayOffset,
	}

	if d.LensID.Valid {
		media.LensID = d.LensID.Int64
	}

	return media
}

func newMedia(media dbMedia) models.Media {
	return media.ToModel()
}

func newDBMedia(media models.Media) dbMedia {
	m := dbMedia{
		ID: media.ID,

		Kind:                    media.Kind,
		Make:                    media.Make,
		Model:                   media.Model,
		Lens:                    media.Lens,
		FocalLength:             media.FocalLength,
		TakenAt:                 media.TakenAt,
		FNumber:                 media.FNumber,
		ExposureTimeNumerator:   media.ExposureTimeNumerator,
		ExposureTimeDenominator: media.ExposureTimeDenominator,
		ISOSpeed:                media.ISOSpeed,
		Latitude:                media.Latitude,
		Longitude:               media.Longitude,
		Altitude:                media.Altitude,

		CreatedAt: media.CreatedAt,
		UpdatedAt: media.UpdatedAt,

		DeviceID: media.DeviceID,

		InstagramCode: media.InstagramCode,

		UTCCorrect: media.UTCCorrect,

		Width:  media.Width,
		Height: media.Height,

		Orientation: media.Orientation,

		DisplayOffset: media.DisplayOffset,
	}

	m.LensID = sql.NullInt64{
		// default to setting to null unless set
		Valid: false,
	}
	if media.LensID != 0 {
		m.LensID = sql.NullInt64{
			Valid: true,
			Int64: media.LensID,
		}
	}

	return m
}

// UpdateMedias is not implemented as a single SQL query since update many in
// place is not supported by goqu and it wasn't worth the work (TODO).
// MediaRepository provides media-specific database operations.
type MediaRepository struct {
	*BaseRepository[models.Media, dbMedia]
}

// NewMediaRepository creates a new media repository instance.
func NewMediaRepository(db *sql.DB) *MediaRepository {
	return &MediaRepository{
		BaseRepository: NewBaseRepository(db, "medias", newMedia, newDBMedia, "taken_at"),
	}
}

// FindByInstagramCode finds media by Instagram code.
func (r *MediaRepository) FindByInstagramCode(ctx context.Context, code string) ([]models.Media, error) {
	return r.FindByField(ctx, "instagram_code", code)
}

// AllDescending retrieves all media ordered by taken_at and created_at descending.
func (r *MediaRepository) AllDescending(ctx context.Context) ([]models.Media, error) {
	var dbMedias []dbMedia

	goquDB := goqu.New("postgres", r.db)
	query := goquDB.From(goqu.T(r.tableName).Schema(r.schema)).
		Select("*").
		Order(goqu.I("taken_at").Desc(), goqu.I("created_at").Desc()).
		Executor()

	err := query.ScanStructsContext(ctx, &dbMedias)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select medias")
	}

	// this is needed in case there are no items added, we don't want to return
	// nil but rather an empty slice
	results := make([]models.Media, 0, len(dbMedias))
	for i := range dbMedias {
		results = append(results, newMedia(dbMedias[i]))
	}

	return results, nil
}

// Legacy function wrappers for backward compatibility with test files.
// These should be removed after all tests are updated.

// CreateMedias creates multiple medias using the repository.
func CreateMedias(ctx context.Context, db *sql.DB, medias []models.Media) ([]models.Media, error) {
	repo := NewMediaRepository(db)
	return repo.Create(ctx, medias)
}

// FindMediasByID finds medias by their IDs using the repository.
func FindMediasByID(ctx context.Context, db *sql.DB, ids []int) ([]models.Media, error) {
	if len(ids) == 0 {
		return []models.Media{}, nil
	}

	var dbMedias []dbMedia

	goquDB := goqu.New("postgres", db)
	query := goquDB.From("photos.medias").Select("*").Where(goqu.Ex{"id": ids}).Executor()
	err := query.ScanStructsContext(ctx, &dbMedias)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select medias by id")
	}

	results := make([]models.Media, 0, len(dbMedias))
	for i := range dbMedias {
		results = append(results, newMedia(dbMedias[i]))
	}

	return results, nil
}

// FindMediasByInstagramCode finds media by Instagram code using the repository.
func FindMediasByInstagramCode(ctx context.Context, db *sql.DB, code string) ([]models.Media, error) {
	repo := NewMediaRepository(db)
	return repo.FindByInstagramCode(ctx, code)
}

// AllMedias gets all medias using the repository.
func AllMedias(ctx context.Context, db *sql.DB, descending bool) ([]models.Media, error) {
	repo := NewMediaRepository(db)
	if descending {
		return repo.AllDescending(ctx)
	}
	return repo.All(ctx)
}

// DeleteMedias deletes medias using the repository.
func DeleteMedias(ctx context.Context, db *sql.DB, medias []models.Media) error {
	if len(medias) == 0 {
		return nil
	}

	ids := make([]int, len(medias))
	for i := range medias {
		ids[i] = medias[i].ID
	}

	goquDB := goqu.New("postgres", db)
	query := goquDB.From("photos.medias").Where(goqu.Ex{"id": ids}).Delete()
	_, err := query.Executor().ExecContext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to delete medias")
	}

	return nil
}

// UpdateMedias updates medias using the repository.
func UpdateMedias(ctx context.Context, db *sql.DB, medias []models.Media) ([]models.Media, error) {
	return BulkUpdateGeneric(ctx, db, "photos.medias", medias, newDBMedia)
}
