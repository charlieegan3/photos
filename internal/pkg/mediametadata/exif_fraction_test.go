package mediametadata

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/require"
)

func TestFractionToDecimal(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fraction       Fraction
		expectedResult float64
		expectError    string
	}{
		"simple example": {
			fraction: Fraction{
				Numerator:   5,
				Denominator: 2,
			},
			expectedResult: 2.5,
		},
		"zero denominator": {
			fraction: Fraction{
				Numerator:   5,
				Denominator: 0,
			},
			expectError: "fraction with 0 denominator cannot be converted to decimal",
		},
		"infinite sequence": {
			fraction: Fraction{
				Numerator:   1,
				Denominator: 3,
			},
			expectedResult: 0.3333333333333333,
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			t.Parallel()

			result, err := testCase.fraction.ToDecimal()

			if testCase.expectError != "" {
				require.Equal(t, err.Error(), testCase.expectError)
			} else {
				require.NoError(t, err)
			}

			td.Cmp(t, result, testCase.expectedResult)
		})
	}
}
