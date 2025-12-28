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

// Ball physics constants
const (
	InitialOrbitRadius   = 0.95  // Ratio of wheel radius for starting orbit
	FinalOrbitRadius     = 0.76  // Ratio where ball enters slots
	InitialAngularSpeed  = 0.45  // Initial ball speed (radians per frame) - fast casino feel
	FrictionCoefficient  = 0.006 // How fast the ball slows down (proportional to speed increase)
	DropThreshold        = 0.03  // Speed at which ball starts dropping
	BounceDecay          = 0.6   // Energy retained after bounce
	SettleThreshold      = 0.001 // Speed at which ball settles
	NumSlots             = 38
	SlotAngle            = 2 * math.Pi / NumSlots

	// Deflector collision constants
	DeflectorAngleHitZone    = 0.12  // Angular hit zone (radians) - slightly larger than visual
	DeflectorRadiusHitZone   = 0.04  // Radial hit zone - how close ball needs to be
	DeflectorCooldownFrames  = 8     // Frames before same deflector can be hit again
	DeflectorSpeedMultiplier = 0.9   // Base speed retention on deflector hit
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

	// Timing
	TicksSinceStart int
	SpinDuration    int // Target spin duration in ticks (frames)

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

// Update updates the ball physics
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
	// Apply friction
	friction := FrictionCoefficient * (1 + float64(b.TicksSinceStart)/float64(b.SpinDuration))
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

	// Check if it's time to drop
	progress := float64(b.TicksSinceStart) / float64(b.SpinDuration)
	if progress > 0.5 && math.Abs(b.AngularSpeed) < DropThreshold*2 {
		b.Phase = PhaseDropping
		b.RadialSpeed = -0.002
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

	// Continue angular motion with more friction
	b.AngularSpeed *= 0.995

	// Update angle
	b.Angle += b.AngularSpeed
	for b.Angle >= 2*math.Pi {
		b.Angle -= 2 * math.Pi
	}
	for b.Angle < 0 {
		b.Angle += 2 * math.Pi
	}

	// Accelerate inward
	b.RadialSpeed -= 0.0003
	b.Radius += b.RadialSpeed

	// Check for deflector collisions
	b.checkDeflectorCollisions(wheelRotation)

	// Check if ball has reached the slot area
	if b.Radius <= FinalOrbitRadius {
		b.Radius = FinalOrbitRadius
		b.Phase = PhaseBouncing
		b.LastBouncePos = b.Angle

		if b.OnBounce != nil {
			b.OnBounce()
		}
	}
}

// updateBouncing handles the ball bouncing between slots
func (b *Ball) updateBouncing(wheelRotation float64) {
	// Reduce speed with each bounce
	b.AngularSpeed *= 0.98

	// Update angle
	b.Angle += b.AngularSpeed
	for b.Angle >= 2*math.Pi {
		b.Angle -= 2 * math.Pi
	}
	for b.Angle < 0 {
		b.Angle += 2 * math.Pi
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

// applyDeflectorRicochet applies chaotic ricochet physics when ball hits a deflector
func (b *Ball) applyDeflectorRicochet(isVerticalDeflector bool) {
	// Randomize the reflection behavior for chaos
	// Base reflection: reverse or partially reverse angular velocity
	reflectionType := randomInt(3)

	switch reflectionType {
	case 0:
		// Full reversal with randomization
		b.AngularSpeed = -b.AngularSpeed * (0.6 + randomFloat()*0.5)
	case 1:
		// Partial reversal - ball continues but slower and offset
		b.AngularSpeed = b.AngularSpeed * (0.4 + randomFloat()*0.3)
	case 2:
		// Strong reversal - ball bounces back harder
		b.AngularSpeed = -b.AngularSpeed * (0.8 + randomFloat()*0.4)
	}

	// Add random angular perturbation (±15-30 degrees worth)
	perturbation := (randomFloat() - 0.5) * 0.5 // ±0.25 radians (~15 degrees)
	b.AngularSpeed += perturbation * 0.1

	// Vertical deflectors (pointing radially) tend to bounce ball more sideways
	// Horizontal deflectors (pointing tangentially) tend to redirect more
	if isVerticalDeflector {
		// More likely to reverse direction
		if randomFloat() > 0.4 {
			b.AngularSpeed = -math.Abs(b.AngularSpeed) * (0.5 + randomFloat()*0.5)
			if randomFloat() > 0.5 {
				b.AngularSpeed = -b.AngularSpeed
			}
		}
	} else {
		// Horizontal: can send ball in either direction
		if randomFloat() > 0.5 {
			b.AngularSpeed = -b.AngularSpeed
		}
	}

	// Speed multiplier: ball can gain or lose energy (0.7-1.1x)
	speedMult := DeflectorSpeedMultiplier + (randomFloat()-0.5)*0.4
	b.AngularSpeed *= speedMult

	// Slightly affect radial speed - deflector can slow the drop or speed it up
	radialPerturbation := (randomFloat() - 0.5) * 0.002
	b.RadialSpeed += radialPerturbation

	// Occasionally, a hard hit can push ball outward slightly
	if randomFloat() > 0.8 {
		b.RadialSpeed += 0.001 // Small outward push
	}

	// Ensure ball doesn't get stuck - always maintain minimum inward motion
	if b.RadialSpeed > -0.0005 {
		b.RadialSpeed = -0.0005
	}

	// Ensure ball has some angular velocity to continue moving
	if math.Abs(b.AngularSpeed) < 0.005 {
		b.AngularSpeed = 0.01
		if randomFloat() > 0.5 {
			b.AngularSpeed = -b.AngularSpeed
		}
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
