package ball

import (
	"math"
	"testing"
)

// fairnessSpins gives ~525 expected hits per slot, enough for the chi-square
// test to detect small biases while still running in about a second.
const fairnessSpins = 20000

func TestFairness(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping statistical fairness simulation in -short mode")
	}

	counts := SimulateSpins(fairnessSpins)

	total := 0
	for _, c := range counts {
		total += c
	}
	if total != fairnessSpins {
		t.Fatalf("%d of %d spins settled", total, fairnessSpins)
	}

	chiSquare := ChiSquare(counts)
	t.Logf("chi-square = %.2f (critical %.2f at p<0.001)", chiSquare, ChiSquareCritical001)
	if !IsFair(counts) {
		t.Errorf("slot distribution is biased: chi-square %.2f > %.2f\ncounts per slot: %v",
			chiSquare, ChiSquareCritical001, counts)
	}
}

func TestChiSquare(t *testing.T) {
	var uniform, skewed, empty [NumSlots]int
	for i := range uniform {
		uniform[i] = 10
	}
	// All hits in one slot is maximally biased: (N-1) * total.
	skewed[0] = NumSlots * 10

	tests := []struct {
		name     string
		counts   [NumSlots]int
		wantChi  float64
		wantFair bool
	}{
		{"uniform", uniform, 0, true},
		{"skewed", skewed, float64((NumSlots - 1) * NumSlots * 10), false},
		{"empty", empty, math.Inf(1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ChiSquare(tt.counts); got != tt.wantChi {
				t.Errorf("ChiSquare = %v, want %v", got, tt.wantChi)
			}
			if got := IsFair(tt.counts); got != tt.wantFair {
				t.Errorf("IsFair = %v, want %v", got, tt.wantFair)
			}
		})
	}
}
