// American Roulette Wheel - A fullscreen roulette wheel for casino night events
package main

import (
	"bytes"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/gobold"

	"roulette-wheel/audio"
	"roulette-wheel/ball"
	"roulette-wheel/stats"
	"roulette-wheel/wheel"
)

// Game constants
const (
	DefaultWindowWidth  = 1280
	DefaultWindowHeight = 720
	TargetFPS           = 60
	StatsPanelWidth     = 300
	WheelSpinSpeed      = 0.05     // Faster initial spin
	WheelFriction       = 0.00008  // Normal friction while ball is spinning
	WheelBrakeFriction  = 0.0002   // Higher friction after ball settles (~2 sec to stop)
)

// Game represents the main game state
type Game struct {
	wheel  *wheel.Wheel
	ball   *ball.Ball
	stats  *stats.Stats
	audio  *audio.Audio

	fontFace *text.GoTextFace

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
}

// NewGame creates a new game instance
func NewGame() *Game {
	g := &Game{
		screenWidth:  DefaultWindowWidth,
		screenHeight: DefaultWindowHeight,
	}

	// Initialize font
	g.initFont()

	// Calculate initial layout
	wheelRadius := g.calculateWheelRadius()
	wheelCenterX, wheelCenterY := g.calculateWheelCenter(wheelRadius)

	// Initialize components
	g.wheel = wheel.New(wheelCenterX, wheelCenterY, wheelRadius)
	g.wheel.FontFace = g.fontFace
	g.ball = ball.New(wheelCenterX, wheelCenterY, wheelRadius)
	g.audio = audio.New()

	// Initialize stats panel
	statsPanelX := float64(g.screenWidth) - StatsPanelWidth
	g.stats = stats.New(statsPanelX, 0, StatsPanelWidth, float64(g.screenHeight))

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

// initFont initializes the font for text rendering
func (g *Game) initFont() {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(gobold.TTF))
	if err != nil {
		log.Printf("Failed to load font: %v", err)
		return
	}

	g.fontFace = &text.GoTextFace{
		Source: s,
		Size:   18,
	}
}

// calculateWheelRadius returns the optimal wheel radius for current screen
func (g *Game) calculateWheelRadius() float64 {
	// Available width for wheel (excluding stats panel)
	availableWidth := float64(g.screenWidth) - StatsPanelWidth
	availableHeight := float64(g.screenHeight)

	// Use the smaller dimension with padding
	padding := 40.0
	maxDiameter := math.Min(availableWidth, availableHeight) - padding*2

	return maxDiameter / 2
}

// calculateWheelCenter returns the center position for the wheel
func (g *Game) calculateWheelCenter(radius float64) (float64, float64) {
	availableWidth := float64(g.screenWidth) - StatsPanelWidth
	centerX := availableWidth / 2
	centerY := float64(g.screenHeight) / 2
	return centerX, centerY
}

// Update handles game logic
func (g *Game) Update() error {
	// Handle input
	g.handleInput()

	// Update wheel rotation
	g.updateWheel()

	// Update ball physics
	g.ball.Update(g.wheel.Rotation, g.wheel.AngularSpeed)

	return nil
}

// handleInput processes user input
func (g *Game) handleInput() {
	// Spin: Click, Space, or Enter
	if !g.isSpinning {
		shouldSpin := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
			inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
			inpututil.IsKeyJustPressed(ebiten.KeyEnter)

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
	}

	// Reset stats: R
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.stats.Reset()
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

	// Start wheel spinning
	g.wheelInitialSpeed = WheelSpinSpeed * (0.8 + float64(randomByte())/255.0*0.4)
	g.wheel.StartSpin(g.wheelInitialSpeed)

	// Start ball spinning (opposite direction)
	g.ball.StartSpin(g.wheel.Rotation)
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

	// Get the winning number from the ball's settled position
	winningNumber := g.ball.GetWinningNumber(wheel.NumberSequence)
	if winningNumber != "" {
		g.stats.RecordResult(winningNumber)
		g.audio.PlayChime()
	}
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
	g.stats.Draw(screen, g.fontFace)

	// Draw controls help (bottom left)
	g.drawControls(screen)

	// Draw mute indicator if muted
	if g.audio.IsMuted() {
		g.drawMuteIndicator(screen)
	}
}

// drawControls draws the control hints at the bottom of the screen
func (g *Game) drawControls(screen *ebiten.Image) {
	if g.fontFace == nil {
		return
	}

	controls := "Click/Space: Spin | F: Fullscreen | M: Mute | R: Reset | Esc: Exit"

	op := &text.DrawOptions{}
	op.GeoM.Translate(20, float64(g.screenHeight)-30)
	op.ColorScale.ScaleWithColor(colorTextDim)
	text.Draw(screen, controls, g.fontFace, op)
}

// drawMuteIndicator shows when audio is muted
func (g *Game) drawMuteIndicator(screen *ebiten.Image) {
	if g.fontFace == nil {
		return
	}

	op := &text.DrawOptions{}
	op.GeoM.Translate(20, 30)
	op.ColorScale.ScaleWithColor(colorMuteIndicator)
	text.Draw(screen, "MUTED", g.fontFace, op)
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

	// Update stats panel
	statsPanelX := float64(g.screenWidth) - StatsPanelWidth
	g.stats.UpdatePosition(statsPanelX, 0, StatsPanelWidth, float64(g.screenHeight))
}

// Color definitions
var (
	colorCasinoFelt    = color.RGBA{0, 80, 40, 255}
	colorTextDim       = color.RGBA{150, 150, 150, 255}
	colorMuteIndicator = color.RGBA{255, 100, 100, 255}
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
