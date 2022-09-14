package activity

import (
	"encoding/json"
	"github.com/charlieegan3/photos/internal/pkg/models"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestParseActivity(t *testing.T) {
	testCases := map[string]struct {
		InputFile              string
		ExpectedPointsJSONFile string
		ExpectedActivity       *models.Activity
	}{
		"FIT wahoo cycle": {
			InputFile:              "./fixtures/wahoo_cycle.fit",
			ExpectedPointsJSONFile: "./data/wahoo_cycle.json",
			ExpectedActivity: &models.Activity{
				Title:       "Cycling",
				Description: "Recorded on WahooFitness ELEMNT ROAM",
				StartTime:   time.Date(2022, 4, 24, 5, 6, 39, 0, time.UTC),
				EndTime:     time.Date(2022, 4, 24, 8, 30, 49, 0, time.UTC),
			},
		},
		"FIT garmin run outdoor": {
			InputFile:              "./fixtures/garmin_outdoor_run.fit",
			ExpectedPointsJSONFile: "./data/garmin_outdoor_run.json",
			ExpectedActivity: &models.Activity{
				Title:       "Running",
				Description: "Recorded on Garmin Forerunner 745",
				StartTime:   time.Date(2022, 4, 30, 8, 1, 16, 0, time.UTC),
				EndTime:     time.Date(2022, 4, 30, 8, 21, 12, 0, time.UTC),
			},
		},
		"FIT garmin run indoor (no points)": {
			InputFile:              "./fixtures/garmin_indoor_run.fit",
			ExpectedPointsJSONFile: "./data/indoor.json",
			ExpectedActivity: &models.Activity{
				Title:       "Treadmill Running",
				Description: "Recorded on Garmin Forerunner 745",
				StartTime:   time.Date(2022, 3, 28, 11, 19, 22, 0, time.UTC),
				EndTime:     time.Date(2022, 3, 28, 11, 49, 25, 0, time.UTC),
			},
		},
		"FIT zwift indoor cycle (no points)": {
			InputFile:              "./fixtures/zwift_indoor_cycle.fit",
			ExpectedPointsJSONFile: "./data/indoor.json",
			ExpectedActivity: &models.Activity{
				Title:       "Indoor Cycling",
				Description: "Recorded on Zwift Unknown Device",
				StartTime:   time.Date(2020, 1, 12, 23, 0, 17, 0, time.UTC),
				EndTime:     time.Date(2020, 1, 12, 23, 40, 6, 0, time.UTC),
			},
		},
		"FIT trainerroad indoor cycle": {
			InputFile:              "./fixtures/trainerroad.fit",
			ExpectedPointsJSONFile: "./data/indoor.json",
			ExpectedActivity: &models.Activity{
				Title:       "Indoor Cycling",
				Description: "Recorded on TrainerRoad Unknown Device",
				StartTime:   time.Date(2022, 5, 11, 6, 5, 17, 0, time.UTC),
				EndTime:     time.Date(2022, 5, 11, 7, 5, 17, 0, time.UTC),
			},
		},

		"GPX strava export": {
			InputFile:              "./fixtures/strava_export.gpx",
			ExpectedPointsJSONFile: "./data/strava_export.json",
			ExpectedActivity: &models.Activity{
				Title:       "Run with Anna",
				Description: "Created by StravaGPX",
				StartTime:   time.Date(2022, 5, 10, 16, 48, 28, 0, time.UTC),
				EndTime:     time.Date(2022, 5, 10, 17, 22, 59, 0, time.UTC),
			},
		},
		"GPX strava iphone": {
			InputFile:              "./fixtures/strava_iphone.gpx",
			ExpectedPointsJSONFile: "./data/strava_iphone.json",
			ExpectedActivity: &models.Activity{
				Title:       "Morning Run",
				Description: "Created by StravaGPX iPhone",
				StartTime:   time.Date(2015, 11, 1, 8, 22, 19, 0, time.UTC),
				EndTime:     time.Date(2015, 11, 1, 8, 48, 58, 0, time.UTC),
			},
		},
		"GPX strava android": {
			InputFile:              "./fixtures/strava_android.gpx",
			ExpectedPointsJSONFile: "./data/strava_android.json",
			ExpectedActivity: &models.Activity{
				Title:       "Been a while...",
				Description: "Created by StravaGPX Android",
				StartTime:   time.Date(2019, 10, 2, 6, 37, 01, 0, time.UTC),
				EndTime:     time.Date(2019, 10, 2, 7, 04, 23, 0, time.UTC),
			},
		},
		"GPX nike plus export": {
			InputFile:              "./fixtures/nikeplusexport.gpx",
			ExpectedPointsJSONFile: "./data/nikeplusexport.json",
			ExpectedActivity: &models.Activity{
				Title:       "2014-02-06T12:43:53Z",
				Description: "Created by Nike+ to GPX exporter (@mccaig)",
				StartTime:   time.Date(2014, 2, 6, 12, 43, 53, 0, time.UTC),
				EndTime:     time.Date(2014, 2, 6, 13, 14, 56, 0, time.UTC),
			},
		},
		"GPX matt stuehler": {
			InputFile:              "./fixtures/matt_stuehler.gpx",
			ExpectedPointsJSONFile: "./data/matt_stuehler.json",
			ExpectedActivity: &models.Activity{
				Title:       "Nike run on Saturday, 11/21/2015 9:30am",
				Description: "Created by MattStuehler.com",
				StartTime:   time.Date(2015, 11, 21, 9, 30, 11, 0, time.UTC),
				EndTime:     time.Date(2015, 11, 21, 9, 49, 42, 0, time.UTC),
			},
		},

		"TCX trainerroad": {
			InputFile:              "./fixtures/trainerroad.tcx",
			ExpectedPointsJSONFile: "./data/indoor.json",
			ExpectedActivity: &models.Activity{
				Title:       "Indoor Cycling",
				Description: "Created by TrainerRoad",
				StartTime:   time.Date(2022, 1, 26, 13, 32, 9, 0, time.UTC),
				EndTime:     time.Date(2022, 1, 26, 15, 2, 9, 0, time.UTC),
			},
		},
		"TCX running": {
			InputFile:              "./fixtures/running.tcx",
			ExpectedPointsJSONFile: "./data/indoor.json",
			ExpectedActivity: &models.Activity{
				Title:       "Running",
				Description: "Created by Unknown Creator",
				StartTime:   time.Date(2015, 6, 27, 8, 02, 38, 0, time.UTC),
				EndTime:     time.Date(2015, 6, 27, 8, 20, 47, 0, time.UTC),
			},
		},
		"TCX kinomap": {
			InputFile:              "./fixtures/kinomap.tcx",
			ExpectedPointsJSONFile: "./data/indoor.json",
			ExpectedActivity: &models.Activity{
				Title:       "Indoor Cycling",
				Description: "Created by KinomapVirtualRide",
				StartTime:   time.Date(2020, 12, 05, 17, 47, 22, 0, time.UTC),
				EndTime:     time.Date(2020, 12, 05, 17, 59, 53, 980000000, time.UTC),
			},
		},
		"TCX nike": {
			InputFile:              "./fixtures/nike.tcx",
			ExpectedPointsJSONFile: "./data/niketcx.json",
			ExpectedActivity: &models.Activity{
				Title:       "Running",
				Description: "Created by Unknown Creator",
				StartTime:   time.Date(2015, 6, 1, 11, 3, 18, 0, time.UTC),
				EndTime:     time.Date(2015, 6, 1, 15, 21, 16, 0, time.UTC),
			},
		},
	}

	for testCaseName, testCaseData := range testCases {
		t.Run(testCaseName, func(t *testing.T) {
			var expectedPoints []Point
			expectedPointsJSONBytes, err := os.ReadFile(testCaseData.ExpectedPointsJSONFile)
			require.NoError(t, err)

			err = json.Unmarshal(expectedPointsJSONBytes, &expectedPoints)
			require.NoError(t, err)

			resultActivity, resultPoints, err := ParseActivity(testCaseData.InputFile)
			require.NoError(t, err)

			td.Cmp(t, resultPoints, expectedPoints)

			if testCaseData.ExpectedActivity != nil {
				td.Cmp(t, &resultActivity, testCaseData.ExpectedActivity)
			}
		})
	}
}
