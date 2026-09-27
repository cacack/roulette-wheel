package wheel

import (
	"image/color"
	"math"
	"strconv"
	"testing"
)

func TestNumberSequence(t *testing.T) {
	if len(NumberSequence) != NumSlots {
		t.Fatalf("len(NumberSequence) = %d, want %d", len(NumberSequence), NumSlots)
	}

	want := map[string]bool{"00": true}
	for n := 0; n <= 36; n++ {
		want[strconv.Itoa(n)] = true
	}
	seen := make(map[string]bool)
	for i, num := range NumberSequence {
		if !want[num] {
			t.Errorf("slot %d: unexpected number %q", i, num)
		}
		if seen[num] {
			t.Errorf("slot %d: duplicate number %q", i, num)
		}
		seen[num] = true
	}
	if len(seen) != len(want) {
		t.Errorf("sequence has %d distinct numbers, want %d", len(seen), len(want))
	}
}

func TestRedNumbers(t *testing.T) {
	if len(RedNumbers) != 18 {
		t.Errorf("len(RedNumbers) = %d, want 18", len(RedNumbers))
	}
	for num := range RedNumbers {
		if IsZero(num) {
			t.Errorf("zero %q marked red", num)
		}
	}
}

// On an American wheel, colors alternate red/black between the green zeros.
func TestColorsAlternateAroundWheel(t *testing.T) {
	for i, num := range NumberSequence {
		next := NumberSequence[(i+1)%len(NumberSequence)]
		if IsZero(num) || IsZero(next) {
			continue
		}
		if IsRed(num) == IsRed(next) {
			t.Errorf("adjacent slots %d (%s) and %d (%s) share a color", i, num, (i+1)%len(NumberSequence), next)
		}
	}
}

func TestNumberClassification(t *testing.T) {
	tests := []struct {
		num   string
		color color.RGBA
		zero  bool
		red   bool
		even  bool
		low   bool
		high  bool
	}{
		{"0", ColorGreen, true, false, false, false, false},
		{"00", ColorGreen, true, false, false, false, false},
		{"1", ColorRed, false, true, false, true, false},
		{"2", ColorBlack, false, false, true, true, false},
		{"10", ColorBlack, false, false, true, true, false},
		{"18", ColorRed, false, true, true, true, false},
		{"19", ColorRed, false, true, false, false, true},
		{"28", ColorBlack, false, false, true, false, true},
		{"35", ColorBlack, false, false, false, false, true},
		{"36", ColorRed, false, true, true, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.num, func(t *testing.T) {
			if got := GetNumberColor(tt.num); got != tt.color {
				t.Errorf("GetNumberColor = %v, want %v", got, tt.color)
			}
			if got := IsZero(tt.num); got != tt.zero {
				t.Errorf("IsZero = %v, want %v", got, tt.zero)
			}
			if got := IsRed(tt.num); got != tt.red {
				t.Errorf("IsRed = %v, want %v", got, tt.red)
			}
			if got := IsEven(tt.num); got != tt.even {
				t.Errorf("IsEven = %v, want %v", got, tt.even)
			}
			if got := IsLow(tt.num); got != tt.low {
				t.Errorf("IsLow = %v, want %v", got, tt.low)
			}
			if got := IsHigh(tt.num); got != tt.high {
				t.Errorf("IsHigh = %v, want %v", got, tt.high)
			}
		})
	}
}

func TestInitialSpeed(t *testing.T) {
	tests := []struct {
		r    float64
		want float64
	}{
		{0, SpinSpeed * 0.8},
		{0.5, SpinSpeed},
		{1, SpinSpeed * 1.2},
	}
	for _, tt := range tests {
		if got := InitialSpeed(tt.r); math.Abs(got-tt.want) > 1e-12 {
			t.Errorf("InitialSpeed(%v) = %v, want %v", tt.r, got, tt.want)
		}
	}
}

func TestTick(t *testing.T) {
	tests := []struct {
		name         string
		speed        float64
		friction     float64
		wantSpeed    float64
		wantRotation float64
	}{
		{"friction slows the wheel", 0.03, 0.01, 0.02, 0.02},
		{"speed never goes negative", 0.005, 0.01, 0, 0},
		{"stopped wheel stays put", 0, 0.01, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := New(0, 0, 100)
			w.StartSpin(tt.speed)
			w.Tick(tt.friction)
			if math.Abs(w.AngularSpeed-tt.wantSpeed) > 1e-12 {
				t.Errorf("speed = %v, want %v", w.AngularSpeed, tt.wantSpeed)
			}
			if math.Abs(w.Rotation-tt.wantRotation) > 1e-12 {
				t.Errorf("rotation = %v, want %v", w.Rotation, tt.wantRotation)
			}
		})
	}
}

func TestGetSlotAngle(t *testing.T) {
	w := New(0, 0, 100)
	for _, rotation := range []float64{0, 1.25, -3} {
		w.Rotation = rotation
		for _, slot := range []int{0, 1, 19, 37} {
			want := float64(slot)*SlotAngle + rotation
			if got := w.GetSlotAngle(slot); math.Abs(got-want) > 1e-12 {
				t.Errorf("rotation %v slot %d: angle = %v, want %v", rotation, slot, got, want)
			}
		}
	}
}

func TestGetDeflectorInfo(t *testing.T) {
	const rotation = 0.5
	deflectors := GetDeflectorInfo(rotation)
	if len(deflectors) != NumDeflectors {
		t.Fatalf("len = %d, want %d", len(deflectors), NumDeflectors)
	}
	spacing := 2 * math.Pi / NumDeflectors
	for i, d := range deflectors {
		if want := float64(i)*spacing + rotation; math.Abs(d.Angle-want) > 1e-12 {
			t.Errorf("deflector %d angle = %v, want %v", i, d.Angle, want)
		}
		if d.RadiusRatio != DeflectorRadiusRatio {
			t.Errorf("deflector %d radius ratio = %v, want %v", i, d.RadiusRatio, DeflectorRadiusRatio)
		}
		if d.IsVertical != (i%2 == 0) {
			t.Errorf("deflector %d IsVertical = %v, want %v", i, d.IsVertical, i%2 == 0)
		}
	}
}
