package activity

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/philhofer/tcx"
	"github.com/tkrajina/gpxgo/gpx"
	"github.com/tormoder/fit"
	"io/ioutil"
	"strings"
)

func ParseActivity(inputFile string) ([]Point, error) {
	points := []Point{}

	rawData, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return points, fmt.Errorf("failed to read file: %w", err)
	}

	parts := strings.Split(inputFile, ".")
	fileType := parts[len(parts)-1]

	switch fileType {
	case "fit":
		return parseFit(rawData)
	case "gpx":
		return parseGPX(rawData)
	case "tcx":
		return parseTCX(rawData)
	default:
		return points, fmt.Errorf("unknown format for input file: %s", fileType)
	}
}

func parseTCX(rawData []byte) ([]Point, error) {
	points := []Point{}
	db := new(tcx.TCXDB)
	err := xml.Unmarshal(rawData, db)
	if err != nil {
		return points, fmt.Errorf("failed to parse TCX data: %w", err)
	}

	for _, activity := range db.Acts.Act {
		if activity.Creator.Name == "KinomapVirtualRide" ||
			activity.Creator.Name == "TrainerRoad" {
			continue
		}
		for _, lap := range activity.Laps {
			for _, trackpoint := range lap.Trk.Pt {
				if trackpoint.Lat == 0 || trackpoint.Long == 0 {
					continue
				}
				points = append(points, Point{
					Timestamp: trackpoint.Time,
					Latitude:  trackpoint.Lat,
					Longitude: trackpoint.Long,
					Altitude:  trackpoint.Alt,
					Velocity:  trackpoint.Speed,
				})
			}
		}
	}

	return points, nil
}

func parseGPX(rawData []byte) ([]Point, error) {
	points := []Point{}

	gpxData, err := gpx.ParseBytes(rawData)
	if err != nil {
		return points, fmt.Errorf("failed to parse GPX data: %w", err)
	}

	for _, track := range gpxData.Tracks {
		for _, segment := range track.Segments {
			for _, point := range segment.Points {
				points = append(points, Point{
					Timestamp:        point.Timestamp.UTC(),
					Latitude:         point.Latitude,
					Longitude:        point.Longitude,
					Altitude:         point.Elevation.Value(),
					Accuracy:         point.HorizontalDilution.Value(),
					VerticalAccuracy: point.VerticalDilution.Value(),
				})
			}
		}
	}

	return points, nil
}

func parseFit(rawData []byte) ([]Point, error) {
	points := []Point{}

	fitData, err := fit.Decode(bytes.NewReader(rawData))
	if err != nil {
		return points, fmt.Errorf("failed to parse data: %w", err)
	}

	// Only accept gpx activities from Wahoo and Garmin
	if fitData.FileId.Manufacturer != fit.ManufacturerGarmin &&
		fitData.FileId.Manufacturer != fit.ManufacturerWahooFitness {
		return points, nil
	}

	activity, err := fitData.Activity()
	if err != nil {
		return points, fmt.Errorf("failed to get activity from fit file: %w", err)
	}

	for _, r := range activity.Records {
		if r.PositionLat.Invalid() || r.PositionLong.Invalid() {
			continue
		}
		points = append(points, Point{
			Timestamp:        r.Timestamp.UTC(),
			Latitude:         r.PositionLat.Degrees(),
			Longitude:        r.PositionLong.Degrees(),
			Altitude:         float64(r.Altitude),
			Accuracy:         float64(r.GpsAccuracy),
			VerticalAccuracy: float64(r.GpsAccuracy),
			Velocity:         float64(r.Speed),
		})
	}

	return points, nil
}
