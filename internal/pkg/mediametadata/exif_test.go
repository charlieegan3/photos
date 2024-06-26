package mediametadata

import (
	"os"
	"testing"
	"time"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/require"
)

func TestExtract(t *testing.T) {
	testCases := map[string]struct {
		sourceFile       string
		expectedMetadata Metadata
	}{
		"iphone jpg": {
			sourceFile: "./samples/iphone-11-pro-max.jpg",
			expectedMetadata: Metadata{
				Make:        "Apple",
				Model:       "iPhone 11 Pro Max",
				Lens:        "iPhone 11 Pro Max back triple camera 6mm f/2",
				FocalLength: "52mm in 35mm format",
				DateTime:    time.Date(2021, time.November, 9, 8, 33, 11, 0, time.UTC),
				FNumber: Fraction{
					Numerator: 2, Denominator: 1,
				},
				ExposureTime: Fraction{
					Numerator:   1,
					Denominator: 122,
				},
				ISOSpeed: 100,
				Altitude: Altitude{
					Value: Fraction{
						Numerator:   1605603,
						Denominator: 16384,
					},
					Ref: 0,
				},
				Latitude: Coordinate{
					Degrees: Fraction{
						Numerator:   51,
						Denominator: 1,
					},
					Minutes: Fraction{
						Numerator:   33,
						Denominator: 1,
					},
					Seconds: Fraction{
						Numerator:   3410,
						Denominator: 100,
					},
					Ref: "N",
				},
				Longitude: Coordinate{
					Degrees: Fraction{
						Numerator:   0,
						Denominator: 1,
					},
					Minutes: Fraction{
						Numerator:   10,
						Denominator: 1,
					},
					Seconds: Fraction{
						Numerator:   707,
						Denominator: 100,
					},
					Ref: "W",
				},
				Width:  4032,
				Height: 3024,
			},
		},
		"iphone jpg non utc": {
			sourceFile: "./samples/iphone-11-pro-max-non-utc.JPG",
			expectedMetadata: Metadata{
				Make:        "Apple",
				Model:       "iPhone 11 Pro Max",
				Lens:        "iPhone 11 Pro Max back triple camera 4.25mm f/1.8",
				FocalLength: "26mm in 35mm format",
				DateTime:    time.Date(2022, time.July, 31, 19, 32, 04, 0, time.UTC),
				FNumber: Fraction{
					Numerator: 9, Denominator: 5,
				},
				ExposureTime: Fraction{
					Numerator:   1,
					Denominator: 33,
				},
				ISOSpeed: 640,
				Altitude: Altitude{
					Value: Fraction{
						Numerator:   24,
						Denominator: 1,
					},
					Ref: 0,
				},
				Latitude: Coordinate{
					Degrees: Fraction{
						Numerator:   51,
						Denominator: 1,
					},
					Minutes: Fraction{
						Numerator:   30,
						Denominator: 1,
					},
					Seconds: Fraction{
						Numerator:   546,
						Denominator: 100,
					},
					Ref: "N",
				},
				Longitude: Coordinate{
					Degrees: Fraction{
						Numerator:   0,
						Denominator: 1,
					},
					Minutes: Fraction{
						Numerator:   26,
						Denominator: 1,
					},
					Seconds: Fraction{
						Numerator:   4924,
						Denominator: 100,
					},
					Ref: "W",
				},
				Width:  3024,
				Height: 4032,
			},
		},
		"fuji jpg": {
			sourceFile: "./samples/x100f.jpg",
			expectedMetadata: Metadata{
				Make:        "FUJIFILM",
				Model:       "X100F",
				Lens:        "FUJINON single focal length lens",
				FocalLength: "23mm (35mm in 35mm format)",
				DateTime:    time.Date(2021, time.November, 13, 15, 38, 02, 0, time.UTC),
				FNumber: Fraction{
					Numerator: 56, Denominator: 10,
				},
				ExposureTime: Fraction{
					Numerator:   1,
					Denominator: 1500,
				},
				ISOSpeed: 400,
				Width:    6000,
				Height:   4000,
			},
		},
		"xt20 with lens": {
			sourceFile: "./samples/xt20-with-lens.jpg",
			expectedMetadata: Metadata{
				Make:        "FUJIFILM",
				Model:       "X-T20",
				Lens:        "XC16-50mmF3.5-5.6 OIS II",
				FocalLength: "68mm in 35mm format",
				DateTime:    time.Date(2022, time.April, 21, 12, 2, 30, 0, time.UTC),
				FNumber: Fraction{
					Numerator: 56, Denominator: 10,
				},
				ExposureTime: Fraction{
					Numerator:   1,
					Denominator: 70,
				},
				ISOSpeed: 1600,
				Width:    1717,
				Height:   1717,
			},
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			b, err := os.ReadFile(testCase.sourceFile)
			require.NoError(t, err)

			metadata, err := ExtractMetadata(b)
			require.NoError(t, err)

			td.Cmp(t, metadata, testCase.expectedMetadata)
		})
	}
}
