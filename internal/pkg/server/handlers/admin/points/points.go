package points

import (
	"database/sql"
	_ "embed"
	"fmt"
	"net/http"
	"time"

	bq "cloud.google.com/go/bigquery"
	"github.com/gobuffalo/plush"
	"github.com/tkrajina/gpxgo/gpx"

	"github.com/charlieegan3/photos/internal/pkg/bigquery"
	"github.com/charlieegan3/photos/internal/pkg/server/templating"
)

//go:embed templates/periodIndex.html.plush
var periodIndexTemplate string

func BuildIndexHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		ctx := plush.NewContext()

		err := renderer(ctx, periodIndexTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}

func BuildPeriodGPXHandler(
	client *bq.Client,
	dataset string,
	table string,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		fromValues, ok := r.URL.Query()["from"]
		if !ok || len(fromValues) != 1 {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("from param required"))
			return
		}

		fromTime, err := time.Parse("2006-01-02", fromValues[0])
		if err != nil {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid from date format"))
			return
		}

		toValues, ok := r.URL.Query()["to"]
		if !ok || len(toValues) != 1 {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("to param required"))
			return
		}

		toTime, err := time.Parse("2006-01-02", toValues[0])
		if err != nil {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid to date format"))
			return
		}
		// make it to the end of the day
		toTime = toTime.Add(24 * time.Hour).Add(-time.Second)

		points, err := bigquery.PointsInRange(r.Context(), client, dataset, table, fromTime, toTime)
		if err != nil {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		segment := gpx.GPXTrackSegment{}

		for _, point := range points {
			alt := gpx.NewNullableFloat64(point.Altitude)
			gpxPoint := gpx.GPXPoint{
				Point: gpx.Point{
					Latitude:  point.Latitude,
					Longitude: point.Longitude,
					Elevation: *alt,
				},
				Timestamp: point.CreatedAt,
			}
			segment.Points = append(segment.Points, gpxPoint)
		}

		g := &gpx.GPX{
			Version: "1.0",
			Creator: "photos.charlieegan3.com",
			Tracks: []gpx.GPXTrack{
				{
					Name:     fmt.Sprintf("%s to %s", fromValues[0], toValues[0]),
					Segments: []gpx.GPXTrackSegment{segment},
				},
			},
		}

		bytes, err := g.ToXml(gpx.ToXmlParams{
			Version: "1.0",
			Indent:  true,
		})

		w.Header().Set("Content-Type", "application/gpx+xml")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s-to-%s.gpx", fromValues[0], toValues[0]))

		w.Write(bytes)
	}
}
