package activity

import (
	"encoding/json"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestParseActivity(t *testing.T) {
	testCases := map[string]struct {
		InputFile              string
		ExpectedPointsJSONFile string
	}{
		"FIT wahoo cycle": {
			InputFile:              "./fixtures/wahoo_cycle.fit",
			ExpectedPointsJSONFile: "./data/wahoo_cycle.json",
		},
		"FIT garmin run outdoor": {
			InputFile:              "./fixtures/garmin_outdoor_run.fit",
			ExpectedPointsJSONFile: "./data/garmin_outdoor_run.json",
		},
		"FIT garmin run indoor (no points)": {
			InputFile:              "./fixtures/garmin_indoor_run.fit",
			ExpectedPointsJSONFile: "./data/indoor.json",
		},
		"FIT zwift indoor cycle (no points)": {
			InputFile:              "./fixtures/zwift_indoor_cycle.fit",
			ExpectedPointsJSONFile: "./data/indoor.json",
		},
		"FIT trainerroad indoor cycle": {
			InputFile:              "./fixtures/trainerroad.fit",
			ExpectedPointsJSONFile: "./data/indoor.json",
		},

		"GPX strava export": {
			InputFile:              "./fixtures/strava_export.gpx",
			ExpectedPointsJSONFile: "./data/strava_export.json",
		},
		"GPX strava iphone": {
			InputFile:              "./fixtures/strava_iphone.gpx",
			ExpectedPointsJSONFile: "./data/strava_iphone.json",
		},
		"GPX strava android": {
			InputFile:              "./fixtures/strava_android.gpx",
			ExpectedPointsJSONFile: "./data/strava_android.json",
		},
		"GPX nike plus export": {
			InputFile:              "./fixtures/nikeplusexport.gpx",
			ExpectedPointsJSONFile: "./data/nikeplusexport.json",
		},
		"GPX matt stuehler": {
			InputFile:              "./fixtures/matt_stuehler.gpx",
			ExpectedPointsJSONFile: "./data/matt_stuehler.json",
		},

		"TCX trainerroad": {
			InputFile:              "./fixtures/trainerroad.tcx",
			ExpectedPointsJSONFile: "./data/indoor.json",
		},
		"TCX running": {
			InputFile:              "./fixtures/running.tcx",
			ExpectedPointsJSONFile: "./data/indoor.json",
		},
		"TCX kinomap": {
			InputFile:              "./fixtures/kinomap.tcx",
			ExpectedPointsJSONFile: "./data/indoor.json",
		},
		"TCX nike": {
			InputFile:              "./fixtures/nike.tcx",
			ExpectedPointsJSONFile: "./data/niketcx.json",
		},
	}

	for testCaseName, testCaseData := range testCases {
		t.Run(testCaseName, func(t *testing.T) {
			var expectedPoints []Point
			expectedPointsJSONBytes, err := os.ReadFile(testCaseData.ExpectedPointsJSONFile)
			require.NoError(t, err)

			err = json.Unmarshal(expectedPointsJSONBytes, &expectedPoints)
			require.NoError(t, err)

			result, err := ParseActivity(testCaseData.InputFile)
			require.NoError(t, err)

			td.Cmp(t, expectedPoints, result)
		})
	}
}
