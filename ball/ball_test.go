package ball

import (
	"math"
	"testing"

	"roulette-wheel/wheel"
)

func TestNumSlotsMatchesWheel(t *testing.T) {
	if NumSlots != wheel.NumSlots || NumSlots != len(wheel.NumberSequence) {
		t.Errorf("ball.NumSlots = %d, wheel.NumSlots = %d, len(NumberSequence) = %d; must match",
			NumSlots, wheel.NumSlots, len(wheel.NumberSequence))
	}
}

func TestSettleMapsAngleToSlot(t *testing.T) {
	quarter := SlotAngle / 4
	tests := []struct {
		name          string
		angle         float64
		wheelRotation float64
		wantSlot      int
	}{
		{"slot 0 center", 0, 0, 0},
		{"slot 5 center", 5 * SlotAngle, 0, 5},
		{"last slot center", 37 * SlotAngle, 0, 37},
		{"with wheel rotation", 1.0 + 10*SlotAngle, 1.0, 10},
		{"negative angle wraps", -SlotAngle, 0, 37},
		{"angle past 2pi wraps", 2*math.Pi + 3*SlotAngle, 0, 3},
		{"multiple turns of rotation", 7*SlotAngle + 4*math.Pi, 4 * math.Pi, 7},
		{"just below half slot rounds down", 12*SlotAngle + 2*quarter - 1e-9, 0, 12},
		{"just above half slot rounds up", 12*SlotAngle + 2*quarter + 1e-9, 0, 13},
		{"near 2pi rounds to slot 0", 2*math.Pi - quarter, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(0, 0, 100)
			b.Phase = PhaseBouncing
			b.Angle = tt.angle
			b.settle(tt.wheelRotation)

			if got := b.GetSettledSlot(); got != tt.wantSlot {
				t.Errorf("settled slot = %d, want %d", got, tt.wantSlot)
			}
			wantAngle := float64(tt.wantSlot)*SlotAngle + tt.wheelRotation
			if math.Abs(b.Angle-wantAngle) > 1e-9 {
				t.Errorf("angle = %v, want snapped to %v", b.Angle, wantAngle)
			}
			if b.AngularSpeed != 0 || b.RadialSpeed != 0 {
				t.Errorf("speeds = (%v, %v), want zero", b.AngularSpeed, b.RadialSpeed)
			}
		})
	}
}

func TestIdleBallDoesNotMove(t *testing.T) {
	b := New(0, 0, 100)
	if b.Phase != PhaseIdle {
		t.Fatalf("new ball phase = %v, want PhaseIdle", b.Phase)
	}
	angle := b.Angle
	b.Update(1.0, 0.02)
	if b.Phase != PhaseIdle || b.Angle != angle || b.TicksSinceStart != 0 {
		t.Errorf("idle ball changed on Update: phase=%v angle=%v ticks=%d", b.Phase, b.Angle, b.TicksSinceStart)
	}
	if b.IsSpinning() || b.IsSettled() {
		t.Error("idle ball reports spinning or settled")
	}
	if got := b.GetWinningNumber(wheel.NumberSequence); got != "" {
		t.Errorf("idle ball winning number = %q, want empty", got)
	}
}

func TestSpinLifecycle(t *testing.T) {
	const wheelSpeed = 0.02
	b := New(0, 0, 100)
	settleCalls := 0
	b.OnSettle = func() { settleCalls++ }

	wheelRotation := 0.0
	b.StartSpin(wheelRotation)
	if b.Phase != PhaseOrbiting {
		t.Fatalf("phase after StartSpin = %v, want PhaseOrbiting", b.Phase)
	}
	if !b.IsSpinning() {
		t.Fatal("IsSpinning = false after StartSpin")
	}

	seen := map[Phase]bool{PhaseOrbiting: true}
	prev := b.Phase
	for iter := 0; iter < maxSpinTicks && !b.IsSettled(); iter++ {
		wheelRotation += wheelSpeed
		b.Update(wheelRotation, wheelSpeed)
		if b.Phase < prev || b.Phase > prev+1 {
			t.Fatalf("tick %d: phase jumped %v -> %v", iter, prev, b.Phase)
		}
		prev = b.Phase
		seen[b.Phase] = true
	}

	if !b.IsSettled() {
		t.Fatalf("ball did not settle within %d ticks (phase %v)", maxSpinTicks, b.Phase)
	}
	for _, p := range []Phase{PhaseOrbiting, PhaseDropping, PhaseBouncing, PhaseSettled} {
		if !seen[p] {
			t.Errorf("phase %v never observed", p)
		}
	}
	if settleCalls != 1 {
		t.Errorf("OnSettle called %d times, want 1", settleCalls)
	}

	slot := b.GetSettledSlot()
	if slot < 0 || slot >= NumSlots {
		t.Fatalf("settled slot %d out of range", slot)
	}
	if got := b.GetWinningNumber(wheel.NumberSequence); got != wheel.NumberSequence[slot] {
		t.Errorf("winning number = %q, want %q", got, wheel.NumberSequence[slot])
	}

	// A settled ball stays in its slot as the wheel keeps turning.
	wheelRotation += 1.5
	b.Update(wheelRotation, wheelSpeed)
	if want := float64(slot)*SlotAngle + wheelRotation; math.Abs(b.Angle-want) > 1e-9 {
		t.Errorf("settled ball angle = %v, want %v (tracking wheel)", b.Angle, want)
	}
	if b.GetSettledSlot() != slot || settleCalls != 1 {
		t.Error("settled ball changed slot or re-settled after further updates")
	}

	b.Reset()
	if b.Phase != PhaseIdle || b.IsSettled() {
		t.Errorf("phase after Reset = %v, want PhaseIdle", b.Phase)
	}
}

func TestDeflectorCollision(t *testing.T) {
	deflectorRadius := wheel.GetDeflectorRadiusRatio()
	// Retention 0.5-0.85 with ±5% perturbation bounds the post-hit speed.
	const minRetention, maxRetention = 0.5 * 0.95, 0.85 * 1.05
	tests := []struct {
		name        string
		speed       float64
		radius      float64
		wantHit     bool
		minAbsSpeed float64
		maxAbsSpeed float64
	}{
		{"counter-clockwise hit keeps direction", -0.05, deflectorRadius, true, 0.05 * minRetention, 0.05 * maxRetention},
		{"clockwise hit keeps direction", 0.05, deflectorRadius, true, 0.05 * minRetention, 0.05 * maxRetention},
		{"slow hit enforces minimum speed", -0.001, deflectorRadius, true, 0.003, 0.003},
		{"outside radial zone misses", -0.05, deflectorRadius + 2*DeflectorRadiusHitZone, false, 0.05, 0.05},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(0, 0, 100)
			b.Phase = PhaseDropping
			b.Angle = 0 // Deflector 0 sits at angle 0 when wheel rotation is 0
			b.Radius = tt.radius
			b.AngularSpeed = tt.speed
			bounces := 0
			b.OnBounce = func() { bounces++ }

			b.checkDeflectorCollisions(0)

			if hit := b.DeflectorHitCount == 1; hit != tt.wantHit {
				t.Fatalf("hit = %v, want %v", hit, tt.wantHit)
			}
			if got := math.Abs(b.AngularSpeed); got < tt.minAbsSpeed-1e-12 || got > tt.maxAbsSpeed+1e-12 {
				t.Errorf("|speed| = %v, want in [%v, %v]", got, tt.minAbsSpeed, tt.maxAbsSpeed)
			}
			if math.Signbit(b.AngularSpeed) != math.Signbit(tt.speed) {
				t.Errorf("speed direction flipped: %v -> %v", tt.speed, b.AngularSpeed)
			}
			if !tt.wantHit {
				return
			}
			if bounces != 1 {
				t.Errorf("OnBounce called %d times, want 1", bounces)
			}
			if b.DeflectorHitCooldowns[0] != DeflectorCooldownFrames {
				t.Errorf("cooldown = %d, want %d", b.DeflectorHitCooldowns[0], DeflectorCooldownFrames)
			}
			if b.RadialSpeed > -0.0003 {
				t.Errorf("radial speed = %v, want inward motion <= -0.0003", b.RadialSpeed)
			}

			// A second check while on cooldown must not register another hit.
			b.checkDeflectorCollisions(0)
			if b.DeflectorHitCount != 1 || bounces != 1 {
				t.Errorf("hit during cooldown: count=%d bounces=%d", b.DeflectorHitCount, bounces)
			}
		})
	}
}
