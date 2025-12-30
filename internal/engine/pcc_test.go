package engine

import (
	"math"
	"testing"
)

func Test_PCC(t *testing.T) {
	x := []float64{1, 2, 3, 4, 5}
	y := []float64{2, 4, 6, 8, 10}

	expectedPCC := 1.0
	calculatedPCC := PCC(x, y, len(x))

	if math.Abs(calculatedPCC-expectedPCC) > 1e-9 {
		t.Errorf("Expected PCC: %f, but got: %f", expectedPCC, calculatedPCC)
	}
}

func Test_PCC_NegativeCorrelation(t *testing.T) {
	x := []float64{1, 2, 3, 4, 5}
	y := []float64{10, 8, 6, 4, 2}

	expectedPCC := -1.0
	calculatedPCC := PCC(x, y, len(x))

	if math.Abs(calculatedPCC-expectedPCC) > 1e-9 {
		t.Errorf("Expected PCC: %f, but got: %f", expectedPCC, calculatedPCC)
	}
}

func Test_PCC_NoCorrelation(t *testing.T) {
	x := []float64{1, 2, 3, 4, 5}
	y := []float64{1, 2, 3, 2, 1}

	expectedPCC := 0.0
	calculatedPCC := PCC(x, y, len(x))

	if math.Abs(calculatedPCC-expectedPCC) > 1e-9 {
		t.Errorf("Expected PCC: %f, but got: %f", expectedPCC, calculatedPCC)
	}
}
