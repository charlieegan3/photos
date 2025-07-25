package mediametadata

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/require"
)

func TestCoordinateToDecimal(t *testing.T) {
	t.Parallel()
	
	testCases := map[string]struct {
		coordinate     Coordinate
		expectedResult float64
		expectError    string
	}{
		"simple example": {
			coordinate: Coordinate{
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
			expectedResult: 51.55947222222222,
		},
		"west/south example": {
			coordinate: Coordinate{
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
				Ref: "W",
			},
			expectedResult: -51.55947222222222,
		},
		"error example": {
			coordinate: Coordinate{
				Degrees: Fraction{
					Numerator:   51,
					Denominator: 0, // this is an error
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
			expectError: "coordinate can't be converted to decimal: fraction with 0 denominator cannot be converted to decimal",
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			t.Parallel()
			
			result, err := testCase.coordinate.ToDecimal()

			if testCase.expectError != "" {
				require.Equal(t, err.Error(), testCase.expectError)
			} else {
				require.NoError(t, err)
			}

			td.Cmp(t, result, testCase.expectedResult)
		})
	}
}
