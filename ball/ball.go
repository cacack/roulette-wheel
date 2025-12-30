// Package ball handles the ball physics and animation for the roulette wheel.
package ball

import (
	"crypto/rand"
	"image/color"
	"math"
	"math/big"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"roulette-wheel/wheel"
)

// Ball physics constants (tuned for 60 TPS fixed timestep)
const (
	InitialOrbitRadius   = 0.95   // Ratio of wheel radius for starting orbit
	FinalOrbitRadius     = 0.76   // Ratio where ball enters slots
	InitialAngularSpeed  = 0.19   // Initial ball speed (radians per tick) - fast casino feel
	FrictionCoefficient  = 0.0013 // How fast the ball slows down
	DropThreshold        = 0.013  // Speed at which ball starts dropping
	BounceDecay          = 0.6    // Energy retained after bounce
	SettleThreshold      = 0.0004 // Speed at which ball settles
	NumSlots             = 38
	SlotAngle            = 2 * math.Pi / NumSlots

	// Deflector collision constants
	DeflectorAngleHitZone    = 0.12 // Angular hit zone (radians) - slightly larger than visual
	DeflectorRadiusHitZone   = 0.04 // Radial hit zone - how close ball needs to be
	DeflectorCooldownFrames  = 8    // Frames before same deflector can be hit again
	DeflectorSpeedMultiplier = 0.9  // Base speed retention on deflector hit
)

// Phase represents the current phase of ball motion
type Phase int

const (
	PhaseIdle Phase = iota
	PhaseOrbiting
	PhaseDropping
	PhaseBouncing
	PhaseSettled
)

// Ball represents the roulette ball state
type Ball struct {
	// Position in polar coordinates relative to wheel center
	Angle  float64 // Current angle in radians
	Radius float64 // Current radius ratio (relative to wheel radius)

	// Velocities
	AngularSpeed float64 // Angular velocity (radians per frame)
	RadialSpeed  float64 // Radial velocity (inward movement)

	// State
	Phase        Phase
	SettledSlot  int     // The slot index where the ball settled (determined by physics, not pre-set)
	WheelRadius  float64 // Reference to the wheel's radius
	WheelCenterX float64
	WheelCenterY float64

	// Bounce physics
	BounceCount   int
	MaxBounces    int
	LastBouncePos float64

	// Deflector collision tracking
	DeflectorHitCooldowns [8]int // Cooldown frames for each deflector (prevents double-hits)
	DeflectorHitCount     int    // Total deflector hits this spin

	// Visual
	Size float64

	// Bounce shadow animation
	BounceHeight     float64 // 0.0 = at surface (just bounced), 1.0 = at peak
	BounceHeightVel  float64 // Velocity of bounce height animation
	ShadowFadeAlpha  float64 // 0.0 = invisible, 1.0 = fully visible (for fade in/out)

	// Timing
	TicksSinceStart    int
	SpinDuration       int     // Target spin duration in ticks (frames)
	TotalAngleTraveled float64 // Total radians traveled during orbit

	// Callbacks
	OnTick   func() // Called when ball passes a slot boundary
	OnBounce func() // Called when ball bounces
	OnSettle func() // Called when ball settles
}

// New creates a new ball
func New(wheelCenterX, wheelCenterY, wheelRadius float64) *Ball {
	return &Ball{
		WheelCenterX: wheelCenterX,
		WheelCenterY: wheelCenterY,
		WheelRadius:  wheelRadius,
		Radius:       InitialOrbitRadius,
		Phase:        PhaseIdle,
		Size:         wheelRadius * 0.035,
		MaxBounces:   5 + randomInt(5), // 5-9 bounces
	}
}

// StartSpin begins the ball spinning
func (b *Ball) StartSpin(wheelRotation float64) {
	// Set initial position with random offset
	b.Angle = wheelRotation + math.Pi + float64(randomInt(360))*math.Pi/180

	// Ball spins opposite to wheel direction
	b.AngularSpeed = -InitialAngularSpeed * (0.9 + randomFloat()*0.2)

	b.Radius = InitialOrbitRadius
	b.RadialSpeed = 0
	b.Phase = PhaseOrbiting
	b.BounceCount = 0
	b.TicksSinceStart = 0
	b.TotalAngleTraveled = 0
	b.MaxBounces = 5 + randomInt(5)
	b.SettledSlot = -1 // Not settled yet

	// Reset deflector collision tracking
	b.DeflectorHitCount = 0
	for i := range b.DeflectorHitCooldowns {
		b.DeflectorHitCooldowns[i] = 0
	}

	// Spin duration approximately 6 seconds (360 frames at 60fps) with small variance
	b.SpinDuration = 330 + randomInt(60) // 330-390 frames (~5.5-6.5 seconds)
}

// Update updates the ball physics (fixed timestep - no dt parameter)
func (b *Ball) Update(wheelRotation, wheelAngularSpeed float64) {
	if b.Phase == PhaseIdle || b.Phase == PhaseSettled {
		if b.Phase == PhaseSettled && b.SettledSlot >= 0 {
			// Keep ball locked to its slot as wheel spins
			slotAngle := float64(b.SettledSlot)*SlotAngle + wheelRotation
			b.Angle = slotAngle
		}
		return
	}

	b.TicksSinceStart++

	switch b.Phase {
	case PhaseOrbiting:
		b.updateOrbiting(wheelRotation)
	case PhaseDropping:
		b.updateDropping(wheelRotation)
	case PhaseBouncing:
		b.updateBouncing(wheelRotation)
	}
}

// updateOrbiting handles the ball spinning around the outer track
func (b *Ball) updateOrbiting(wheelRotation float64) {
	// Track total distance traveled
	b.TotalAngleTraveled += math.Abs(b.AngularSpeed)

	// First 4 laps: constant speed, no friction
	// After 4 laps: friction kicks in
	requiredLaps := 4.0
	if b.TotalAngleTraveled >= requiredLaps*2*math.Pi {
		// Apply friction after 4 laps
		friction := FrictionCoefficient
		if b.AngularSpeed < 0 {
			b.AngularSpeed += friction
			if b.AngularSpeed > -DropThreshold {
				b.AngularSpeed = -DropThreshold
			}
		} else {
			b.AngularSpeed -= friction
			if b.AngularSpeed < DropThreshold {
				b.AngularSpeed = DropThreshold
			}
		}
	}

	// Update position
	oldAngle := b.Angle
	b.Angle += b.AngularSpeed

	// Normalize angle
	for b.Angle >= 2*math.Pi {
		b.Angle -= 2 * math.Pi
	}
	for b.Angle < 0 {
		b.Angle += 2 * math.Pi
	}

	// Check for slot boundary crossing (tick sound)
	if b.OnTick != nil {
		oldSlot := int(oldAngle / SlotAngle)
		newSlot := int(b.Angle / SlotAngle)
		if oldSlot != newSlot {
			b.OnTick()
		}
	}

	// Check if it's time to drop (only after 3 laps and speed is low enough)
	if b.TotalAngleTraveled >= requiredLaps*2*math.Pi && math.Abs(b.AngularSpeed) < DropThreshold*2 {
		b.Phase = PhaseDropping
		b.RadialSpeed = -0.0002 // Gentle initial inward drift
	}
}

// updateDropping handles the ball falling toward the center
func (b *Ball) updateDropping(wheelRotation float64) {
	// Decrement deflector cooldowns
	for i := range b.DeflectorHitCooldowns {
		if b.DeflectorHitCooldowns[i] > 0 {
			b.DeflectorHitCooldowns[i]--
		}
	}

	// Continue angular motion with friction
	b.AngularSpeed *= 0.998

	// Update angle
	b.Angle += b.AngularSpeed
	for b.Angle >= 2*math.Pi {
		b.Angle -= 2 * math.Pi
	}
	for b.Angle < 0 {
		b.Angle += 2 * math.Pi
	}

	// Accelerate inward gradually (gravity-like curve)
	b.RadialSpeed -= 0.00004
	b.Radius += b.RadialSpeed

	// Check for deflector collisions
	b.checkDeflectorCollisions(wheelRotation)

	// Check if ball has reached the slot area
	if b.Radius <= FinalOrbitRadius {
		b.Radius = FinalOrbitRadius
		b.Phase = PhaseBouncing
		b.LastBouncePos = b.Angle

		// Initialize bounce shadow - start at ground level (just impacted)
		b.BounceHeight = 0.0
		b.BounceHeightVel = 0.12 // First bounce is high and dramatic
		b.ShadowFadeAlpha = 1.0  // Shadow fully visible immediately

		if b.OnBounce != nil {
			b.OnBounce()
		}
	}
}

// updateBouncing handles the ball bouncing between slots
func (b *Ball) updateBouncing(wheelRotation float64) {
	// Reduce speed with friction
	b.AngularSpeed *= 0.992

	// Update angle
	b.Angle += b.AngularSpeed
	for b.Angle >= 2*math.Pi {
		b.Angle -= 2 * math.Pi
	}
	for b.Angle < 0 {
		b.Angle += 2 * math.Pi
	}

	// Update bounce height animation (parabolic arc between bounces)
	b.BounceHeight += b.BounceHeightVel
	b.BounceHeightVel -= 0.005 // Gravity pulls height back down

	// Clamp height to valid range
	if b.BounceHeight < 0 {
		b.BounceHeight = 0
	}
	if b.BounceHeight > 1 {
		b.BounceHeight = 1
	}

	// Check if ball has moved enough to bounce
	angleMoved := math.Abs(b.Angle - b.LastBouncePos)
	if angleMoved > math.Pi {
		angleMoved = 2*math.Pi - angleMoved
	}

	bounceThreshold := SlotAngle * (0.5 + randomFloat()*0.5)
	if angleMoved > bounceThreshold && b.BounceCount < b.MaxBounces {
		b.BounceCount++
		b.LastBouncePos = b.Angle

		// Reverse direction with some randomness
		b.AngularSpeed = -b.AngularSpeed * BounceDecay * (0.8 + randomFloat()*0.4)

		// Reset bounce height to ground level (ball just hit surface)
		b.BounceHeight = 0.0
		// Decreasing initial velocity with each bounce (losing energy)
		// Steeper decay: bounces get noticeably smaller
		energyFactor := 1.0 - float64(b.BounceCount)*0.15
		if energyFactor < 0.2 {
			energyFactor = 0.2
		}
		b.BounceHeightVel = 0.12 * energyFactor

		if b.OnBounce != nil {
			b.OnBounce()
		}
	}

	// Check if ball should settle
	if math.Abs(b.AngularSpeed) < SettleThreshold || b.BounceCount >= b.MaxBounces {
		b.settle(wheelRotation)
	}
}

// settle finalizes the ball in the nearest slot based on its current position
func (b *Ball) settle(wheelRotation float64) {
	b.Phase = PhaseSettled

	// Reset shadow state when settling
	b.BounceHeight = 0
	b.BounceHeightVel = 0
	b.ShadowFadeAlpha = 0 // Shadow disappears when settled

	// Calculate which slot the ball is in based on its current angle relative to the wheel
	// Normalize the ball's angle relative to wheel rotation
	relativeAngle := b.Angle - wheelRotation
	for relativeAngle < 0 {
		relativeAngle += 2 * math.Pi
	}
	for relativeAngle >= 2*math.Pi {
		relativeAngle -= 2 * math.Pi
	}

	// Find the nearest slot
	b.SettledSlot = int(math.Floor(relativeAngle/SlotAngle+0.5)) % NumSlots
	if b.SettledSlot < 0 {
		b.SettledSlot += NumSlots
	}

	// Snap to the center of that slot
	slotAngle := float64(b.SettledSlot)*SlotAngle + wheelRotation
	b.Angle = slotAngle
	b.Radius = (FinalOrbitRadius + 0.65) / 2 // Middle of slot area
	b.AngularSpeed = 0
	b.RadialSpeed = 0

	if b.OnSettle != nil {
		b.OnSettle()
	}
}

// checkDeflectorCollisions checks for and handles collisions with deflectors
func (b *Ball) checkDeflectorCollisions(wheelRotation float64) {
	// Get deflector info from wheel package
	deflectors := wheel.GetDeflectorInfo(wheelRotation)
	deflectorRadiusRatio := wheel.GetDeflectorRadiusRatio()

	// Check if ball is in the deflector zone (radius-wise)
	radiusDiff := math.Abs(b.Radius - deflectorRadiusRatio)
	if radiusDiff > DeflectorRadiusHitZone {
		return // Ball is not at deflector radius
	}

	// Check each deflector
	for i, deflector := range deflectors {
		// Skip if this deflector is on cooldown
		if b.DeflectorHitCooldowns[i] > 0 {
			continue
		}

		// Calculate angular distance to deflector
		angleDiff := b.Angle - deflector.Angle
		// Normalize to [-PI, PI]
		for angleDiff > math.Pi {
			angleDiff -= 2 * math.Pi
		}
		for angleDiff < -math.Pi {
			angleDiff += 2 * math.Pi
		}

		// Check if within hit zone
		if math.Abs(angleDiff) < DeflectorAngleHitZone {
			// COLLISION! Apply ricochet physics
			b.applyDeflectorRicochet(deflector.IsVertical)

			// Set cooldown to prevent immediate re-collision
			b.DeflectorHitCooldowns[i] = DeflectorCooldownFrames
			b.DeflectorHitCount++

			// Trigger bounce callback for sound
			if b.OnBounce != nil {
				b.OnBounce()
			}

			// Only process one deflector hit per frame
			break
		}
	}
}

// applyDeflectorRicochet applies realistic ricochet physics when ball hits a deflector
func (b *Ball) applyDeflectorRicochet(isVerticalDeflector bool) {
	// Remember original direction (negative = counter-clockwise, positive = clockwise)
	originalSign := 1.0
	if b.AngularSpeed < 0 {
		originalSign = -1.0
	}

	// Deflector slows the ball (0.5-0.85 of original speed)
	speedRetention := 0.5 + randomFloat()*0.35
	b.AngularSpeed = math.Abs(b.AngularSpeed) * speedRetention

	// Add small random perturbation (±5% of current speed)
	perturbation := b.AngularSpeed * (randomFloat() - 0.5) * 0.1
	b.AngularSpeed += perturbation

	// Restore original direction - ball continues same way, just slower
	b.AngularSpeed *= originalSign

	// Slightly affect radial speed - deflector can slow or speed up the drop
	radialPerturbation := (randomFloat() - 0.5) * 0.0008
	b.RadialSpeed += radialPerturbation

	// Occasionally, a hard hit can push ball outward slightly
	if randomFloat() > 0.85 {
		b.RadialSpeed += 0.0004
	}

	// Ensure ball doesn't get stuck - maintain minimum inward motion
	if b.RadialSpeed > -0.0003 {
		b.RadialSpeed = -0.0003
	}

	// Ensure ball has minimum angular velocity
	if math.Abs(b.AngularSpeed) < 0.003 {
		b.AngularSpeed = 0.003 * originalSign
	}
}

// DrawShadow renders a soft-edged shadow under the ball during bouncing
// The shadow size and opacity are inversely correlated with bounce height:
// - Ball at impact (low): shadow is small and dark
// - Ball at peak (high): shadow is large and light
func (b *Ball) DrawShadow(screen *ebiten.Image) {
	// Only draw shadow during bouncing phase
	if b.Phase != PhaseBouncing || b.ShadowFadeAlpha <= 0 {
		return
	}

	x, y := b.GetPosition()

	// Shadow position is offset down and right from ball (light source upper-left)
	shadowOffsetX := b.Size * 0.3
	shadowOffsetY := b.Size * 0.4

	// Shadow size grows with height (ball appears higher = shadow spreads more)
	// At height 0 (impact): shadow is 1.5x ball size (tight under ball)
	// At height 1 (peak): shadow is 4x ball size (spread out)
	baseShadowSize := b.Size * 1.5
	maxShadowSize := b.Size * 4.0
	shadowSize := baseShadowSize + (maxShadowSize-baseShadowSize)*b.BounceHeight

	// Shadow opacity decreases with height (higher ball = lighter shadow)
	// At height 0 (impact): opacity is 220 (very dark, clearly visible)
	// At height 1 (peak): opacity is 80 (lighter but still visible)
	baseOpacity := 220.0
	minOpacity := 80.0
	opacity := baseOpacity - (baseOpacity-minOpacity)*b.BounceHeight

	// Apply fade alpha for smooth fade in/out
	opacity *= b.ShadowFadeAlpha

	// Draw soft-edged shadow using concentric ellipses with decreasing opacity
	// This creates a gradient/blur effect
	numLayers := 8
	for i := numLayers - 1; i >= 0; i-- {
		// Layer factor: 0.0 for innermost, 1.0 for outermost
		layerFactor := float64(i) / float64(numLayers-1)

		// Each layer is progressively larger
		layerSize := shadowSize * (0.3 + layerFactor*0.7)

		// Each layer is progressively more transparent (outer = more transparent)
		layerOpacity := opacity * (1.0 - layerFactor*0.6)

		// Make shadow slightly elliptical (wider than tall) for realism
		ellipseWidth := float32(layerSize * 1.2)
		ellipseHeight := float32(layerSize * 0.9)

		// Draw ellipse using DrawFilledCircle with scale transform
		// Since Ebitengine doesn't have native ellipse, we use a scaled circle approach
		// For simplicity, draw as a circle (close enough for the effect)
		shadowX := float32(x + shadowOffsetX)
		shadowY := float32(y + shadowOffsetY)

		// Use average of width/height for circle approximation
		avgRadius := (ellipseWidth + ellipseHeight) / 2

		shadowColor := color.RGBA{0, 0, 0, uint8(layerOpacity)}
		vector.DrawFilledCircle(screen, shadowX, shadowY, avgRadius/2, shadowColor, false)
	}
}

// Draw renders the ball
func (b *Ball) Draw(screen *ebiten.Image) {
	if b.Phase == PhaseIdle {
		return
	}

	x, y := b.GetPosition()

	// Ball shadow
	vector.DrawFilledCircle(screen, float32(x+3), float32(y+3), float32(b.Size), color.RGBA{0, 0, 0, 100}, false)

	// Main ball (silver/white)
	vector.DrawFilledCircle(screen, float32(x), float32(y), float32(b.Size), color.RGBA{230, 230, 230, 255}, false)

	// Highlight
	vector.DrawFilledCircle(screen, float32(x-b.Size*0.3), float32(y-b.Size*0.3), float32(b.Size*0.3), color.RGBA{255, 255, 255, 200}, false)
}

// GetPosition returns the ball's current screen position
func (b *Ball) GetPosition() (float64, float64) {
	r := b.Radius * b.WheelRadius

	// During bouncing, offset the ball outward at the apex of each bounce
	// This creates the visual effect of the ball arcing outward as it bounces
	if b.Phase == PhaseBouncing && b.BounceHeight > 0 {
		// Outward offset peaks at bounce apex, scaled by ball size
		// Use a curve that peaks in the middle of the bounce arc
		outwardOffset := b.BounceHeight * b.Size * 1.5
		r += outwardOffset
	}

	x := b.WheelCenterX + r*math.Cos(b.Angle)
	y := b.WheelCenterY + r*math.Sin(b.Angle)
	return x, y
}

// GetWinningNumber returns the number string of the winning slot
func (b *Ball) GetWinningNumber(numberSequence []string) string {
	if b.Phase != PhaseSettled || b.SettledSlot < 0 || b.SettledSlot >= len(numberSequence) {
		return ""
	}
	return numberSequence[b.SettledSlot]
}

// GetSettledSlot returns the slot index where the ball settled (-1 if not settled)
func (b *Ball) GetSettledSlot() int {
	if b.Phase != PhaseSettled {
		return -1
	}
	return b.SettledSlot
}

// IsSettled returns true if the ball has settled into a slot
func (b *Ball) IsSettled() bool {
	return b.Phase == PhaseSettled
}

// IsSpinning returns true if the ball is currently in motion
func (b *Ball) IsSpinning() bool {
	return b.Phase != PhaseIdle && b.Phase != PhaseSettled
}

// Reset resets the ball to idle state
func (b *Ball) Reset() {
	b.Phase = PhaseIdle
	b.Angle = 0
	b.Radius = InitialOrbitRadius
	b.AngularSpeed = 0
	b.RadialSpeed = 0
	b.TicksSinceStart = 0
	b.BounceCount = 0

	// Reset shadow state
	b.BounceHeight = 0
	b.BounceHeightVel = 0
	b.ShadowFadeAlpha = 0
}

// UpdatePosition updates the wheel center position (for window resize)
func (b *Ball) UpdatePosition(centerX, centerY, radius float64) {
	b.WheelCenterX = centerX
	b.WheelCenterY = centerY
	b.WheelRadius = radius
	b.Size = radius * 0.035
}

// Helper functions for randomness using crypto/rand

func randomInt(max int) int {
	if max <= 0 {
		return 0
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0
	}
	return int(n.Int64())
}

func randomFloat() float64 {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return 0.5
	}
	return float64(n.Int64()) / 1000000.0
}
