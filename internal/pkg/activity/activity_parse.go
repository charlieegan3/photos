package activity

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"os"
	"strings"

	"github.com/charlieegan3/photos/internal/pkg/models"
	"github.com/philhofer/tcx"
	"github.com/tkrajina/gpxgo/gpx"
	"github.com/tormoder/fit"
)

func ParseActivity(inputFile string) (models.Activity, []Point, error) {
	points := []Point{}

	rawData, err := os.ReadFile(inputFile)
	if err != nil {
		return models.Activity{}, points, fmt.Errorf("failed to read file: %w", err)
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
		return models.Activity{}, points, fmt.Errorf("unknown format for input file: %s", fileType)
	}
}

func parseTCX(rawData []byte) (models.Activity, []Point, error) {
	activity := models.Activity{}
	points := []Point{}

	db := new(tcx.TCXDB)
	err := xml.Unmarshal(rawData, db)
	if err != nil {
		return activity, points, fmt.Errorf("failed to parse TCX data: %w", err)
	}

	creator := "Unknown Creator"
	sport := "Unknown Sport"
	for _, act := range db.Acts.Act {
		if act.Creator.Name != "" {
			creator = act.Creator.Name
		}

		if act.Sport != "" {
			sport = act.Sport
		}

		for i, lap := range act.Laps {
			for j, trackpoint := range lap.Trk.Pt {
				if i == 0 && j == 0 {
					activity.StartTime = trackpoint.Time.UTC()
				}
				if i == len(act.Laps)-1 && j == (len(lap.Trk.Pt))-1 {
					activity.EndTime = trackpoint.Time.UTC()
				}

				if trackpoint.Lat == 0 || trackpoint.Long == 0 {
					continue
				}

				// don't save points for virtual rides
				if act.Creator.Name == "KinomapVirtualRide" ||
					act.Creator.Name == "TrainerRoad" {
					continue
				}
				points = append(points, Point{
					Timestamp: trackpoint.Time.UTC(),
					Latitude:  trackpoint.Lat,
					Longitude: trackpoint.Long,
					Altitude:  trackpoint.Alt,
					Velocity:  trackpoint.Speed,
				})
			}
		}
	}

	if sport == "Biking" {
		if creator == "TrainerRoad" || creator == "KinomapVirtualRide" {
			sport = "Indoor Cycling"
		}
	}

	activity.Title = sport
	activity.Description = fmt.Sprintf("Created by %s", creator)

	return activity, points, nil
}

func parseGPX(rawData []byte) (models.Activity, []Point, error) {
	activity := models.Activity{}
	points := []Point{}

	gpxData, err := gpx.ParseBytes(rawData)
	if err != nil {
		return activity, points, fmt.Errorf("failed to parse GPX data: %w", err)
	}

	activity.Description = fmt.Sprintf("Created by %s", gpxData.Creator)

	for _, track := range gpxData.Tracks {
		activity.Title = track.Name
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

	if len(points) > 1 {
		activity.StartTime = points[0].Timestamp
		activity.EndTime = points[len(points)-1].Timestamp
	}

	return activity, points, nil
}

func parseFit(rawData []byte) (models.Activity, []Point, error) {
	activity := models.Activity{}
	points := []Point{}

	fitData, err := fit.Decode(bytes.NewReader(rawData))
	if err != nil {
		return activity, points, fmt.Errorf("failed to parse data: %w", err)
	}

	fitData.FileId.Manufacturer.String()

	fitActivity, err := fitData.Activity()
	if err != nil {
		return activity, points, fmt.Errorf("failed to get activity from fit file: %w", err)
	}

	typeName := "Unknown Type"
	activityTypes := []string{}
	for _, s := range fitActivity.Sessions {
		sport := s.Sport.String()
		subSport := s.SubSport.String()

		if subSport != "" && subSport != "Generic" {
			if subSport == "VirtualActivity" {
				subSport = "Indoor"
			}
			if subSport == "IndoorCycling" {
				subSport = "Indoor"
			}
			sport = fmt.Sprintf("%s %s", subSport, sport)
		}

		activityTypes = append(activityTypes, sport)
	}
	if len(activityTypes) > 0 {
		typeName = strings.Join(activityTypes, ", ")
	}

	deviceName := "Unknown Device"
	for _, d := range fitActivity.DeviceInfos {
		if d.DeviceIndex == 0 {
			if d.ProductName != "" {
				deviceName = d.ProductName
			}
			// garmin forerunner 745 watch
			if d.GetProduct() == fit.GarminProduct(3589) {
				deviceName = "Forerunner 745"
			}
			break
		}
	}

	activity.Title = cases.Title(language.English).String(typeName)
	activity.Description = fmt.Sprintf("Recorded on %s %s", fitData.FileId.Manufacturer.String(), deviceName)

	for i, r := range fitActivity.Records {
		if i == 0 {
			activity.StartTime = r.Timestamp
		}
		if i == len(fitActivity.Records)-1 {
			activity.EndTime = r.Timestamp
		}
		if r.PositionLat.Invalid() || r.PositionLong.Invalid() {
			continue
		}
		// Only accept gpx activities from Wahoo and Garmin
		if fitData.FileId.Manufacturer != fit.ManufacturerGarmin &&
			fitData.FileId.Manufacturer != fit.ManufacturerWahooFitness {
			continue
		}
		points = append(points, Point{
			Timestamp:        r.Timestamp,
			Latitude:         r.PositionLat.Degrees(),
			Longitude:        r.PositionLong.Degrees(),
			Altitude:         float64(r.Altitude),
			Accuracy:         float64(r.GpsAccuracy),
			VerticalAccuracy: float64(r.GpsAccuracy),
			Velocity:         float64(r.Speed),
		})
	}

	return activity, points, nil
}
