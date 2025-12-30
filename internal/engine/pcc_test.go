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

func Test_PCC_EmptySlices(t *testing.T) {
	x := []float64{}
	y := []float64{}

	calculatedPCC := PCC(x, y, len(x))

	// Empty slices should return 0.0 (handled by NaN check in PCC function)
	if math.Abs(calculatedPCC) > 1e-9 {
		t.Errorf("Expected PCC: 0.0 for empty slices, but got: %f", calculatedPCC)
	}
}

func Test_PCC_SingleElement(t *testing.T) {
	x := []float64{5.0}
	y := []float64{10.0}

	calculatedPCC := PCC(x, y, len(x))

	// Single element slices should return 0.0 (no variance, division by zero handled)
	if math.Abs(calculatedPCC) > 1e-9 {
		t.Errorf("Expected PCC: 0.0 for single element slices, but got: %f", calculatedPCC)
	}
}

func Test_PCC_TwoElements(t *testing.T) {
	x := []float64{1.0, 2.0}
	y := []float64{3.0, 6.0}

	expectedPCC := 1.0
	calculatedPCC := PCC(x, y, len(x))

	if math.Abs(calculatedPCC-expectedPCC) > 1e-9 {
		t.Errorf("Expected PCC: %f for two elements, but got: %f", expectedPCC, calculatedPCC)
	}
}

func Test_PCC_ZeroVarianceX(t *testing.T) {
	x := []float64{5.0, 5.0, 5.0, 5.0}
	y := []float64{1.0, 2.0, 3.0, 4.0}

	calculatedPCC := PCC(x, y, len(x))

	// Zero variance in x should return 0.0 (division by zero handled by NaN check)
	if math.Abs(calculatedPCC) > 1e-9 {
		t.Errorf("Expected PCC: 0.0 for zero variance in x, but got: %f", calculatedPCC)
	}
}

func Test_PCC_ZeroVarianceY(t *testing.T) {
	x := []float64{1.0, 2.0, 3.0, 4.0}
	y := []float64{5.0, 5.0, 5.0, 5.0}

	calculatedPCC := PCC(x, y, len(x))

	// Zero variance in y should return 0.0 (division by zero handled by NaN check)
	if math.Abs(calculatedPCC) > 1e-9 {
		t.Errorf("Expected PCC: 0.0 for zero variance in y, but got: %f", calculatedPCC)
	}
}

func Test_PCC_ZeroVarianceBoth(t *testing.T) {
	x := []float64{5.0, 5.0, 5.0, 5.0}
	y := []float64{3.0, 3.0, 3.0, 3.0}

	calculatedPCC := PCC(x, y, len(x))

	// Zero variance in both should return 0.0 (division by zero handled by NaN check)
	if math.Abs(calculatedPCC) > 1e-9 {
		t.Errorf("Expected PCC: 0.0 for zero variance in both, but got: %f", calculatedPCC)
	}
}

func Test_PCC_DifferentLengths(t *testing.T) {
	// Test with x shorter than y - PCC iterates over x, so this should work
	x := []float64{1, 2, 3}
	y := []float64{2, 4, 6, 8, 10}

	calculatedPCC := PCC(x, y, len(x))

	// Should calculate correlation based on first len(x) elements
	expectedPCC := 1.0 // First 3 elements are perfectly correlated
	if math.Abs(calculatedPCC-expectedPCC) > 1e-9 {
		t.Errorf("Expected PCC: %f for different length slices, but got: %f", expectedPCC, calculatedPCC)
	}
}
