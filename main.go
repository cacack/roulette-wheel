// American Roulette Wheel - A fullscreen roulette wheel for casino night events
package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"roulette-wheel/audio"
	"roulette-wheel/ball"
	"roulette-wheel/fonts"
	"roulette-wheel/stats"
	"roulette-wheel/wheel"
)

// Game constants
const (
	DefaultWindowWidth  = 1280
	DefaultWindowHeight = 720
	TargetFPS           = 60
	StatsPanelWidth     = 300
	HistoryPanelWidth   = 110  // Left side history panel
	WheelSpinSpeed      = 0.085    // Fast initial spin for casino feel
	WheelFriction       = 0.00014  // Normal friction while ball is spinning (proportional to speed increase)
	WheelBrakeFriction  = 0.00034  // Higher friction after ball settles (~2 sec to stop)
)

// Animation constants
const (
	AnimationHoldFrames   = 120 // ~2 seconds at 60fps
	AnimationShrinkFrames = 60  // ~1 second at 60fps
)

// Animation phases
const (
	AnimPhaseNone = iota
	AnimPhaseHold
	AnimPhaseShrink
)

// debugEntry represents a single log entry during a spin
type debugEntry struct {
	TimeMs     int64   // Milliseconds since spin start
	UpdateNum  int     // Update count
	TPS        float64 // Actual TPS
	FPS        float64 // Actual FPS
	WheelSpeed float64 // Wheel angular speed
	BallSpeed  float64 // Ball angular speed
	BallPhase  string  // Ball phase name
	BallRadius float64 // Ball radius ratio
}

// Game represents the main game state
type Game struct {
	wheel  *wheel.Wheel
	ball   *ball.Ball
	stats  *stats.Stats
	audio  *audio.Audio

	fontMgr *fonts.Manager

	// Window state
	screenWidth  int
	screenHeight int
	isFullscreen bool

	// Game state
	isSpinning        bool
	spinStarted       bool
	wheelInitialSpeed float64
	lastTickSlot      int
	ballSettled       bool // True when ball has settled but wheel is still spinning
	resultDeclared    bool // True when the final result has been declared

	// Winning animation state
	animPhase        int     // Current animation phase (AnimPhaseNone, AnimPhaseHold, AnimPhaseShrink)
	animFrameCount   int     // Frame counter for current phase
	animWinningNum   string  // The winning number to display
	animWinningColor color.RGBA // Color of the winning number

	// Debug mode
	showDebug    bool
	updateCount  int     // Counter for Update() calls during spin
	minTPS       float64 // Min TPS seen during spin
	maxTPS       float64 // Max TPS seen during spin
	debugLog     []debugEntry // Log entries for current spin
	spinStartTime int64  // Unix timestamp when spin started (ms)
}

// NewGame creates a new game instance
func NewGame() *Game {
	g := &Game{
		screenWidth:  DefaultWindowWidth,
		screenHeight: DefaultWindowHeight,
	}

	// Initialize font manager with premium fonts
	g.fontMgr = fonts.NewManager()

	// Calculate initial layout
	wheelRadius := g.calculateWheelRadius()
	wheelCenterX, wheelCenterY := g.calculateWheelCenter(wheelRadius)

	// Initialize components
	g.wheel = wheel.New(wheelCenterX, wheelCenterY, wheelRadius)
	g.wheel.FontSource = g.fontMgr.Source()
	g.ball = ball.New(wheelCenterX, wheelCenterY, wheelRadius)
	g.audio = audio.New()

	// Initialize stats panels (right stats panel, left history panel)
	statsPanelX := float64(g.screenWidth) - StatsPanelWidth
	g.stats = stats.New(
		statsPanelX, 0, StatsPanelWidth, float64(g.screenHeight),
		0, 0, HistoryPanelWidth, float64(g.screenHeight),
	)

	// Set up ball callbacks for audio
	g.ball.OnTick = func() {
		g.audio.PlayTick()
	}
	g.ball.OnBounce = func() {
		g.audio.PlayBounce()
	}
	g.ball.OnSettle = func() {
		g.audio.PlaySettle()
		// Ball has settled, but wheel is still spinning
		// Result will be declared when wheel stops
		g.ballSettled = true
	}

	return g
}


// calculateWheelRadius returns the optimal wheel radius for current screen
func (g *Game) calculateWheelRadius() float64 {
	// Available width for wheel (excluding both panels)
	availableWidth := float64(g.screenWidth) - StatsPanelWidth - HistoryPanelWidth
	availableHeight := float64(g.screenHeight)

	// Use the smaller dimension with padding
	padding := 40.0
	maxDiameter := math.Min(availableWidth, availableHeight) - padding*2

	return maxDiameter / 2
}

// calculateWheelCenter returns the center position for the wheel
func (g *Game) calculateWheelCenter(radius float64) (float64, float64) {
	// Center wheel between the two panels
	availableWidth := float64(g.screenWidth) - StatsPanelWidth - HistoryPanelWidth
	centerX := HistoryPanelWidth + availableWidth/2
	centerY := float64(g.screenHeight) / 2
	return centerX, centerY
}

// Update handles game logic
func (g *Game) Update() error {
	// Track debug data during spin
	if g.isSpinning {
		g.updateCount++
		tps := ebiten.ActualTPS()
		if tps > 0 {
			if tps < g.minTPS {
				g.minTPS = tps
			}
			if tps > g.maxTPS {
				g.maxTPS = tps
			}
		}

		// Log entry every 10 updates to keep log manageable
		if g.showDebug && g.updateCount%10 == 0 {
			g.debugLog = append(g.debugLog, debugEntry{
				TimeMs:     time.Now().UnixMilli() - g.spinStartTime,
				UpdateNum:  g.updateCount,
				TPS:        tps,
				FPS:        ebiten.ActualFPS(),
				WheelSpeed: g.wheel.AngularSpeed,
				BallSpeed:  g.ball.AngularSpeed,
				BallPhase:  g.getBallPhaseName(),
				BallRadius: g.ball.Radius,
			})
		}
	}

	// Handle input
	g.handleInput()

	// Update wheel rotation
	g.updateWheel()

	// Update ball physics
	g.ball.Update(g.wheel.Rotation, g.wheel.AngularSpeed)

	// Sync rolling audio with ball state
	g.syncRollingAudio()

	// Update winning animation
	g.updateWinningAnimation()

	return nil
}

// updateWinningAnimation handles the winning number reveal animation state machine
func (g *Game) updateWinningAnimation() {
	switch g.animPhase {
	case AnimPhaseHold:
		g.animFrameCount++
		if g.animFrameCount >= AnimationHoldFrames {
			// Transition to shrink phase
			g.animPhase = AnimPhaseShrink
			g.animFrameCount = 0
		}

	case AnimPhaseShrink:
		g.animFrameCount++
		if g.animFrameCount >= AnimationShrinkFrames {
			// Animation complete - record the result to stats
			g.stats.RecordResult(g.animWinningNum)
			g.animPhase = AnimPhaseNone
			g.animFrameCount = 0
			g.animWinningNum = ""
		}
	}
}

// syncRollingAudio synchronizes the rolling sound with ball physics
func (g *Game) syncRollingAudio() {
	switch g.ball.Phase {
	case ball.PhaseOrbiting, ball.PhaseDropping:
		// Start rolling sound if not already playing
		if !g.audio.IsRolling() {
			g.audio.StartRolling()
		}
		// Update volume based on angular speed
		speedRatio := math.Abs(g.ball.AngularSpeed) / ball.InitialAngularSpeed
		g.audio.UpdateRollingVolume(speedRatio)

	case ball.PhaseBouncing, ball.PhaseSettled, ball.PhaseIdle:
		// Stop rolling sound when ball is bouncing or settled
		if g.audio.IsRolling() {
			g.audio.StopRolling()
		}
	}
}

// handleInput processes user input
func (g *Game) handleInput() {
	// Spin: Click, Space, or Enter (only when not spinning and animation is done)
	if !g.isSpinning && g.animPhase == AnimPhaseNone {
		shouldSpin := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
			inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
			inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
			inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) ||
			inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) ||
			inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) ||
			inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) ||
			inpututil.IsKeyJustPressed(ebiten.KeyPageDown) ||
			inpututil.IsKeyJustPressed(ebiten.KeyPageUp)

		if shouldSpin {
			g.startSpin()
		}
	}

	// Fullscreen: F or F11
	if inpututil.IsKeyJustPressed(ebiten.KeyF) || inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		g.toggleFullscreen()
	}

	// Mute: M
	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		g.audio.ToggleMute()
		// Stop rolling sound when muted
		if g.audio.IsMuted() {
			g.audio.StopRolling()
		}
	}

	// Reset stats: R
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.resetGame()
	}

	// Debug mode: D
	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.showDebug = !g.showDebug
	}

	// Exit: Escape
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		// In fullscreen, first exit fullscreen; otherwise quit
		if g.isFullscreen {
			g.toggleFullscreen()
		} else {
			// Could implement exit confirmation here
		}
	}
}

// startSpin initiates a new spin
func (g *Game) startSpin() {
	g.isSpinning = true
	g.spinStarted = true
	g.ballSettled = false
	g.resultDeclared = false

	// Reset debug tracking
	g.updateCount = 0
	g.minTPS = 9999
	g.maxTPS = 0
	g.debugLog = nil // Clear previous log
	g.spinStartTime = time.Now().UnixMilli()

	// Reset animation state
	g.animPhase = AnimPhaseNone
	g.animFrameCount = 0
	g.animWinningNum = ""

	// Start wheel spinning
	g.wheelInitialSpeed = WheelSpinSpeed * (0.8 + float64(randomByte())/255.0*0.4)
	g.wheel.StartSpin(g.wheelInitialSpeed)

	// Start ball spinning (opposite direction)
	g.ball.StartSpin(g.wheel.Rotation)
}

// resetGame resets the game state and stops all sounds
func (g *Game) resetGame() {
	// Stop rolling sound first
	g.audio.StopRolling()

	// Reset ball and wheel state
	g.ball.Reset()
	g.wheel.Stop()

	// Reset game flags
	g.isSpinning = false
	g.spinStarted = false
	g.ballSettled = false
	g.resultDeclared = false

	// Reset animation state
	g.animPhase = AnimPhaseNone
	g.animFrameCount = 0
	g.animWinningNum = ""

	// Reset stats
	g.stats.Reset()
}

// updateWheel updates wheel rotation and applies friction
func (g *Game) updateWheel() {
	if !g.isSpinning {
		return
	}

	// Apply friction to wheel (use brake friction after ball settles)
	currentSpeed := g.wheel.AngularSpeed
	if currentSpeed > 0 {
		friction := WheelFriction
		if g.ballSettled {
			friction = WheelBrakeFriction
		}
		currentSpeed -= friction
		if currentSpeed < 0 {
			currentSpeed = 0
		}
	}
	g.wheel.SetSpeed(currentSpeed)
	g.wheel.Update()

	// Check if wheel has stopped (speed is effectively zero)
	wheelStopped := currentSpeed <= 0.0001

	if wheelStopped && g.ballSettled && !g.resultDeclared {
		// Wheel has stopped and ball has settled - declare the result
		g.wheel.Stop()
		g.declareResult()
	}
}

// declareResult calculates and records the final result after wheel stops
func (g *Game) declareResult() {
	g.resultDeclared = true
	g.isSpinning = false
	g.spinStarted = false

	// Write debug log if debug mode is enabled
	if g.showDebug && len(g.debugLog) > 0 {
		g.writeDebugLog()
	}

	// Get the winning number from the ball's settled position
	winningNumber := g.ball.GetWinningNumber(wheel.NumberSequence)
	if winningNumber != "" {
		// Start the winning animation
		g.animWinningNum = winningNumber
		g.animWinningColor = getNumberColor(winningNumber)
		g.animPhase = AnimPhaseHold
		g.animFrameCount = 0
		g.audio.PlayChime()
		// Note: stats.RecordResult will be called when animation completes
	}
}

// getNumberColor returns the color for a roulette number
func getNumberColor(num string) color.RGBA {
	if num == "0" || num == "00" {
		return color.RGBA{0, 128, 0, 255} // Green
	}
	// Red numbers on American roulette wheel
	redNumbers := map[string]bool{
		"1": true, "3": true, "5": true, "7": true, "9": true,
		"12": true, "14": true, "16": true, "18": true, "19": true,
		"21": true, "23": true, "25": true, "27": true, "30": true,
		"32": true, "34": true, "36": true,
	}
	if redNumbers[num] {
		return color.RGBA{185, 30, 30, 255} // Red
	}
	return color.RGBA{25, 25, 25, 255} // Black
}

// toggleFullscreen toggles between windowed and fullscreen modes
func (g *Game) toggleFullscreen() {
	g.isFullscreen = !g.isFullscreen
	ebiten.SetFullscreen(g.isFullscreen)
}

// Draw renders the game
func (g *Game) Draw(screen *ebiten.Image) {
	// Clear screen with dark green casino felt
	screen.Fill(colorCasinoFelt)

	// Draw wheel
	g.wheel.Draw(screen)

	// Draw ball
	g.ball.Draw(screen)

	// Draw stats panel
	g.stats.Draw(screen, g.fontMgr)

	// Draw controls help (bottom left)
	g.drawControls(screen)

	// Draw mute indicator if muted
	if g.audio.IsMuted() {
		g.drawMuteIndicator(screen)
	}

	// Draw winning animation (on top of everything)
	g.drawWinningAnimation(screen)

	// Draw debug info if enabled
	if g.showDebug {
		g.drawDebugInfo(screen)
	}
}

// drawDebugInfo displays TPS, FPS, and speed diagnostics
func (g *Game) drawDebugInfo(screen *ebiten.Image) {
	face := g.fontMgr.Face(fonts.SizeSmall)
	if face == nil {
		return
	}

	// Line 1: Current TPS/FPS
	line1 := fmt.Sprintf("TPS: %.1f (min:%.1f max:%.1f) | FPS: %.1f",
		ebiten.ActualTPS(), g.minTPS, g.maxTPS, ebiten.ActualFPS())

	// Line 2: Speeds and phase
	line2 := fmt.Sprintf("Wheel: %.5f | Ball: %.5f | Phase: %s | Updates: %d",
		g.wheel.AngularSpeed, g.ball.AngularSpeed, g.getBallPhaseName(), g.updateCount)

	// Line 3: Log status
	line3 := fmt.Sprintf("Log entries: %d (saves on spin end when debug enabled)", len(g.debugLog))

	op := &text.DrawOptions{}
	op.GeoM.Translate(20, 50)
	op.ColorScale.ScaleWithColor(color.RGBA{255, 0, 255, 255}) // Magenta
	text.Draw(screen, line1, face, op)

	op2 := &text.DrawOptions{}
	op2.GeoM.Translate(20, 70)
	op2.ColorScale.ScaleWithColor(color.RGBA{255, 0, 255, 255}) // Magenta
	text.Draw(screen, line2, face, op2)

	op3 := &text.DrawOptions{}
	op3.GeoM.Translate(20, 90)
	op3.ColorScale.ScaleWithColor(color.RGBA{200, 0, 200, 255}) // Dimmer magenta
	text.Draw(screen, line3, face, op3)
}

// getBallPhaseName returns a human-readable name for the ball phase
func (g *Game) getBallPhaseName() string {
	switch g.ball.Phase {
	case ball.PhaseIdle:
		return "Idle"
	case ball.PhaseOrbiting:
		return "Orbiting"
	case ball.PhaseDropping:
		return "Dropping"
	case ball.PhaseBouncing:
		return "Bouncing"
	case ball.PhaseSettled:
		return "Settled"
	default:
		return "Unknown"
	}
}

// writeDebugLog writes the collected debug log to a file
func (g *Game) writeDebugLog() {
	if len(g.debugLog) == 0 {
		return
	}

	filename := fmt.Sprintf("debug_spin_%s.log", time.Now().Format("20060102_150405"))
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("Failed to create debug log: %v", err)
		return
	}
	defer f.Close()

	// Write header
	fmt.Fprintf(f, "# Roulette Debug Log\n")
	fmt.Fprintf(f, "# Generated: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(f, "# TPS Range: %.1f - %.1f\n", g.minTPS, g.maxTPS)
	fmt.Fprintf(f, "# Total Updates: %d\n", g.updateCount)
	fmt.Fprintf(f, "#\n")
	fmt.Fprintf(f, "TimeMs\tUpdate\tTPS\tFPS\tWheelSpeed\tBallSpeed\tBallPhase\tBallRadius\n")

	// Write entries
	for _, e := range g.debugLog {
		fmt.Fprintf(f, "%d\t%d\t%.2f\t%.2f\t%.6f\t%.6f\t%s\t%.4f\n",
			e.TimeMs, e.UpdateNum, e.TPS, e.FPS,
			e.WheelSpeed, e.BallSpeed, e.BallPhase, e.BallRadius)
	}

	log.Printf("Debug log saved: %s (%d entries)", filename, len(g.debugLog))
}

// drawControls draws the control hints at the bottom of the screen
func (g *Game) drawControls(screen *ebiten.Image) {
	face := g.fontMgr.Face(fonts.SizeSmall)
	if face == nil {
		return
	}

	controls := "Click/Space/Arrows: Spin | F: Fullscreen | M: Mute | R: Reset | D: Debug | Esc: Exit"

	op := &text.DrawOptions{}
	op.GeoM.Translate(20, float64(g.screenHeight)-30)
	op.ColorScale.ScaleWithColor(colorTextDim)
	text.Draw(screen, controls, face, op)
}

// drawMuteIndicator shows when audio is muted
func (g *Game) drawMuteIndicator(screen *ebiten.Image) {
	face := g.fontMgr.Face(fonts.SizeBody)
	if face == nil {
		return
	}

	op := &text.DrawOptions{}
	op.GeoM.Translate(20, 30)
	op.ColorScale.ScaleWithColor(colorMuteIndicator)
	text.Draw(screen, "MUTED", face, op)
}

// drawWinningAnimation renders the winning number reveal animation
func (g *Game) drawWinningAnimation(screen *ebiten.Image) {
	if g.animPhase == AnimPhaseNone || g.animWinningNum == "" {
		return
	}

	// Full size circle centered in wheel, extending to the number slots
	// SlotInnerRatio = 0.65, so use 0.62 to reach right up to the numbers
	fullRadius := g.wheel.Radius * 0.62

	// Target position: LAST section in left panel
	// The history panel starts at X=0, and chips are centered at X = HistoryPanelWidth/2
	// LAST chip position: 25 (padding) + 30 (title) + 38 (chip radius) = 93
	targetX := float64(HistoryPanelWidth) / 2
	targetY := 93.0        // Matches LAST chip center position
	targetRadius := 38.0   // Same size as LAST section chip

	// Wheel center for full reveal (inside the brown ring)
	centerX := g.wheel.CenterX
	centerY := g.wheel.CenterY

	var currentX, currentY, currentRadius float64

	switch g.animPhase {
	case AnimPhaseHold:
		// Static at wheel center with full size
		currentX = centerX
		currentY = centerY
		currentRadius = fullRadius

	case AnimPhaseShrink:
		// Animate from wheel center to target
		progress := float64(g.animFrameCount) / float64(AnimationShrinkFrames)
		// Ease-out cubic for smooth deceleration
		eased := 1 - math.Pow(1-progress, 3)

		currentX = centerX + (targetX-centerX)*eased
		currentY = centerY + (targetY-centerY)*eased
		currentRadius = fullRadius + (targetRadius-fullRadius)*eased
	}

	// Draw subtle glow behind the circle for premium effect
	glowRadius := currentRadius * 1.15
	glowColor := color.RGBA{g.animWinningColor.R, g.animWinningColor.G, g.animWinningColor.B, 80}
	vector.DrawFilledCircle(screen, float32(currentX), float32(currentY), float32(glowRadius), glowColor, false)

	// Draw the circle with winning color
	vector.DrawFilledCircle(screen, float32(currentX), float32(currentY), float32(currentRadius), g.animWinningColor, false)

	// Draw gold border
	borderWidth := float32(currentRadius * 0.05)
	if borderWidth < 2 {
		borderWidth = 2
	}
	vector.StrokeCircle(screen, float32(currentX), float32(currentY), float32(currentRadius), borderWidth, colorGold, false)

	// Draw the winning number text using properly sized font (no scaling!)
	// Select appropriate font size based on circle radius
	var face *text.GoTextFace
	if currentRadius >= 200 {
		face = g.fontMgr.Face(fonts.SizeWinLarge) // 144pt for large display
	} else if currentRadius >= 100 {
		face = g.fontMgr.Face(fonts.SizeWinSmall) // 96pt
	} else if currentRadius >= 60 {
		face = g.fontMgr.Face(fonts.SizeHuge) // 72pt
	} else if currentRadius >= 40 {
		face = g.fontMgr.Face(fonts.SizeXLarge) // 48pt
	} else if currentRadius >= 25 {
		face = g.fontMgr.Face(fonts.SizeLarge) // 36pt
	} else {
		face = g.fontMgr.Face(fonts.SizeMedium) // 24pt
	}

	if face != nil {
		// Measure text for precise centering
		textWidth, textHeight := text.Measure(g.animWinningNum, face, 0)

		// Draw shadow for depth (offset slightly down and right)
		shadowOffset := currentRadius * 0.02
		if shadowOffset < 1 {
			shadowOffset = 1
		}
		shadowOp := &text.DrawOptions{}
		shadowOp.GeoM.Translate(currentX-textWidth/2+shadowOffset, currentY-textHeight/2+shadowOffset)
		shadowOp.ColorScale.ScaleWithColor(color.RGBA{0, 0, 0, 100})
		text.Draw(screen, g.animWinningNum, face, shadowOp)

		// Draw main text centered in circle
		op := &text.DrawOptions{}
		op.GeoM.Translate(currentX-textWidth/2, currentY-textHeight/2)
		op.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, g.animWinningNum, face, op)
	}
}

// Layout handles window resizing
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if outsideWidth != g.screenWidth || outsideHeight != g.screenHeight {
		g.screenWidth = outsideWidth
		g.screenHeight = outsideHeight
		g.onResize()
	}
	return g.screenWidth, g.screenHeight
}

// onResize updates component positions when window size changes
func (g *Game) onResize() {
	wheelRadius := g.calculateWheelRadius()
	wheelCenterX, wheelCenterY := g.calculateWheelCenter(wheelRadius)

	// Update wheel
	g.wheel.CenterX = wheelCenterX
	g.wheel.CenterY = wheelCenterY
	g.wheel.Radius = wheelRadius

	// Update ball
	g.ball.UpdatePosition(wheelCenterX, wheelCenterY, wheelRadius)

	// Update stats panels (right stats panel, left history panel)
	statsPanelX := float64(g.screenWidth) - StatsPanelWidth
	g.stats.UpdatePosition(
		statsPanelX, 0, StatsPanelWidth, float64(g.screenHeight),
		0, 0, HistoryPanelWidth, float64(g.screenHeight),
	)
}

// Color definitions
var (
	colorCasinoFelt    = color.RGBA{0, 80, 40, 255}
	colorTextDim       = color.RGBA{150, 150, 150, 255}
	colorMuteIndicator = color.RGBA{255, 100, 100, 255}
	colorGold          = color.RGBA{218, 165, 32, 255}
)

// randomByte returns a random byte for variation (using simple PRNG for non-critical randomness)
var prngState uint32 = 42

func randomByte() byte {
	prngState = prngState*1103515245 + 12345
	return byte((prngState >> 16) & 0xFF)
}

func main() {
	ebiten.SetWindowSize(DefaultWindowWidth, DefaultWindowHeight)
	ebiten.SetWindowTitle("American Roulette")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(TargetFPS)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
