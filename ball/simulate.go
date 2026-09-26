package ball

import (
	"math"

	"roulette-wheel/wheel"
)

// ChiSquareCritical001 is the chi-square critical value for 37 degrees of
// freedom (38 slots) at p < 0.001. A statistic above this indicates bias.
const ChiSquareCritical001 = 69.35

// maxSpinTicks is the safety limit on ticks for a single spin to settle
const maxSpinTicks = 10000

// SimulateSpins runs numSpins headless spins and returns hit counts per slot.
// Each spin drives the wheel as the game does: a varied initial speed slowed
// by wheel.Friction. Spins that fail to settle within the safety limit are not
// counted, so callers should compare the total against numSpins.
func SimulateSpins(numSpins int) [NumSlots]int {
	var counts [NumSlots]int
	w := wheel.New(400, 300, 200)

	for i := 0; i < numSpins; i++ {
		w.StartSpin(wheel.InitialSpeed(randomFloat()))
		b := New(w.CenterX, w.CenterY, w.Radius)
		b.StartSpin(w.Rotation)

		for tick := 0; tick < maxSpinTicks && !b.IsSettled(); tick++ {
			w.Tick(wheel.Friction)
			b.Update(w.Rotation, w.AngularSpeed)
		}

		if b.IsSettled() {
			counts[b.GetSettledSlot()]++
		}
	}
	return counts
}

// ChiSquare returns the chi-square statistic of slot counts against a
// uniform distribution. With no hits at all the statistic is undefined, so it
// returns +Inf, which fails any comparison against a critical value.
func ChiSquare(counts [NumSlots]int) float64 {
	total := 0
	for _, c := range counts {
		total += c
	}
	if total == 0 {
		return math.Inf(1)
	}
	expected := float64(total) / NumSlots

	chiSquare := 0.0
	for _, c := range counts {
		diff := float64(c) - expected
		chiSquare += (diff * diff) / expected
	}
	return chiSquare
}

// IsFair reports whether slot counts are consistent with a fair wheel at
// p < 0.001.
func IsFair(counts [NumSlots]int) bool {
	return ChiSquare(counts) <= ChiSquareCritical001
}
