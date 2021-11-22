package media

import (
	"os"
	"testing"

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
				Make:  "Apple",
				Model: "iPhone 11 Pro Max",
				FNumber: Fraction{
					Numerator: 2, Denominator: 1,
				},
				ShutterSpeed: SignedFraction{
					Numerator: 328711, Denominator: 47450,
				},
				ISOSpeed: 100,
			},
		},
		"fuji jpg": {
			sourceFile: "./samples/x100f.jpg",
			expectedMetadata: Metadata{
				Make:  "FUJIFILM",
				Model: "X100F",
				FNumber: Fraction{
					Numerator: 56, Denominator: 10,
				},
				ShutterSpeed: SignedFraction{
					Numerator: 10550747, Denominator: 1000000,
				},
				ISOSpeed: 400,
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
