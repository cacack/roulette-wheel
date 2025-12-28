// Package wheel handles the rendering and layout of the American roulette wheel.
package wheel

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// American roulette wheel number sequence (clockwise from 0)
var NumberSequence = []string{
	"0", "28", "9", "26", "30", "11", "7", "20", "32", "17",
	"5", "22", "34", "15", "3", "24", "36", "13", "1", "00",
	"27", "10", "25", "29", "12", "8", "19", "31", "18", "6",
	"21", "33", "16", "4", "23", "35", "14", "2",
}

// Number colors for American roulette
var RedNumbers = map[string]bool{
	"1": true, "3": true, "5": true, "7": true, "9": true,
	"12": true, "14": true, "16": true, "18": true, "19": true,
	"21": true, "23": true, "25": true, "27": true, "30": true,
	"32": true, "34": true, "36": true,
}

// Color constants
var (
	ColorGreen     = color.RGBA{0, 128, 0, 255}
	ColorRed       = color.RGBA{185, 30, 30, 255}
	ColorBlack     = color.RGBA{25, 25, 25, 255}
	ColorGold      = color.RGBA{218, 165, 32, 255}
	ColorWood      = color.RGBA{139, 90, 43, 255}
	ColorWoodDark  = color.RGBA{101, 67, 33, 255}
	ColorChrome    = color.RGBA{192, 192, 192, 255}
	ColorBallTrack = color.RGBA{60, 40, 25, 255}
)

// Wheel dimensions as ratios of the wheel radius
const (
	BallTrackOuterRatio = 1.0
	BallTrackInnerRatio = 0.92
	SlotOuterRatio      = 0.88
	SlotInnerRatio      = 0.65
	CenterRatio         = 0.30
	DeflectorRatio      = 0.90
	NumSlots            = 38
	SlotAngle           = 2 * math.Pi / NumSlots
)

// Wheel represents the roulette wheel state
type Wheel struct {
	CenterX       float64
	CenterY       float64
	Radius        float64
	Rotation      float64 // Current rotation angle in radians
	AngularSpeed  float64 // Current angular velocity
	IsSpinning    bool
	TargetSlot    int     // The slot where the ball will land
	InitialSpeed  float64 // Initial spinning speed
	FontFace      *text.GoTextFace
}

// New creates a new wheel at the given position and size
func New(centerX, centerY, radius float64) *Wheel {
	return &Wheel{
		CenterX: centerX,
		CenterY: centerY,
		Radius:  radius,
	}
}

// Update updates the wheel state
func (w *Wheel) Update() {
	if w.IsSpinning && w.AngularSpeed > 0 {
		w.Rotation += w.AngularSpeed
		// Keep rotation in [0, 2*PI)
		for w.Rotation >= 2*math.Pi {
			w.Rotation -= 2 * math.Pi
		}
		for w.Rotation < 0 {
			w.Rotation += 2 * math.Pi
		}
	}
}

// StartSpin begins spinning the wheel
func (w *Wheel) StartSpin(initialSpeed float64) {
	w.IsSpinning = true
	w.InitialSpeed = initialSpeed
	w.AngularSpeed = initialSpeed
}

// SetSpeed sets the current angular speed
func (w *Wheel) SetSpeed(speed float64) {
	w.AngularSpeed = speed
}

// Stop stops the wheel
func (w *Wheel) Stop() {
	w.IsSpinning = false
	w.AngularSpeed = 0
}

// GetSlotAngle returns the angle to the center of a specific slot
func (w *Wheel) GetSlotAngle(slotIndex int) float64 {
	return float64(slotIndex)*SlotAngle + w.Rotation
}

// GetSlotPosition returns the world position of a slot center
func (w *Wheel) GetSlotPosition(slotIndex int, radiusRatio float64) (float64, float64) {
	angle := w.GetSlotAngle(slotIndex)
	r := w.Radius * radiusRatio
	x := w.CenterX + r*math.Cos(angle)
	y := w.CenterY + r*math.Sin(angle)
	return x, y
}

// Draw renders the wheel
func (w *Wheel) Draw(screen *ebiten.Image) {
	// Draw ball track (outer wooden rim)
	drawFilledCircle(screen, w.CenterX, w.CenterY, w.Radius*BallTrackOuterRatio, ColorWood)
	drawFilledCircle(screen, w.CenterX, w.CenterY, w.Radius*BallTrackInnerRatio, ColorBallTrack)

	// Draw deflectors (metal diamonds on the ball track)
	w.drawDeflectors(screen)

	// Draw slot area background
	drawFilledCircle(screen, w.CenterX, w.CenterY, w.Radius*SlotOuterRatio, ColorWoodDark)

	// Draw individual slots
	w.drawSlots(screen)

	// Draw center cone/hub
	w.drawCenter(screen)

	// Draw chrome ring separating slots from center
	drawRing(screen, w.CenterX, w.CenterY, w.Radius*SlotInnerRatio, w.Radius*SlotInnerRatio-4, ColorChrome)
}

// drawDeflectors draws the metal deflectors around the ball track
func (w *Wheel) drawDeflectors(screen *ebiten.Image) {
	numDeflectors := 8
	deflectorAngle := 2 * math.Pi / float64(numDeflectors)
	deflectorRadius := w.Radius * DeflectorRatio

	for i := 0; i < numDeflectors; i++ {
		angle := float64(i)*deflectorAngle + w.Rotation
		x := w.CenterX + deflectorRadius*math.Cos(angle)
		y := w.CenterY + deflectorRadius*math.Sin(angle)

		// Draw diamond-shaped deflector
		size := w.Radius * 0.03
		drawDiamond(screen, x, y, size, angle, ColorChrome)
	}
}

// drawSlots draws all 38 numbered slots
func (w *Wheel) drawSlots(screen *ebiten.Image) {
	outerR := w.Radius * SlotOuterRatio
	innerR := w.Radius * SlotInnerRatio

	for i, numStr := range NumberSequence {
		startAngle := float64(i)*SlotAngle + w.Rotation - SlotAngle/2
		endAngle := startAngle + SlotAngle

		// Determine slot color
		var slotColor color.RGBA
		if numStr == "0" || numStr == "00" {
			slotColor = ColorGreen
		} else if RedNumbers[numStr] {
			slotColor = ColorRed
		} else {
			slotColor = ColorBlack
		}

		// Draw slot wedge
		drawWedge(screen, w.CenterX, w.CenterY, innerR, outerR, startAngle, endAngle, slotColor)

		// Draw slot dividers
		dividerAngle := float64(i)*SlotAngle + w.Rotation + SlotAngle/2
		x1 := w.CenterX + innerR*math.Cos(dividerAngle)
		y1 := w.CenterY + innerR*math.Sin(dividerAngle)
		x2 := w.CenterX + outerR*math.Cos(dividerAngle)
		y2 := w.CenterY + outerR*math.Sin(dividerAngle)
		vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 2, ColorChrome, false)

		// Draw number
		w.drawNumber(screen, i, numStr)
	}
}

// drawNumber draws a number on a slot
func (w *Wheel) drawNumber(screen *ebiten.Image, slotIndex int, numStr string) {
	if w.FontFace == nil {
		return
	}

	// Position at middle of slot
	midRadius := w.Radius * (SlotOuterRatio + SlotInnerRatio) / 2
	angle := float64(slotIndex)*SlotAngle + w.Rotation

	x := w.CenterX + midRadius*math.Cos(angle)
	y := w.CenterY + midRadius*math.Sin(angle)

	// Calculate font size based on wheel radius (scale with wheel size)
	fontSize := w.Radius * 0.045
	if fontSize < 8 {
		fontSize = 8
	}
	if fontSize > 24 {
		fontSize = 24
	}

	// Create a scaled font face for the number
	scaledFace := &text.GoTextFace{
		Source: w.FontFace.Source,
		Size:   fontSize,
	}

	// Measure text to center it properly
	textWidth, textHeight := text.Measure(numStr, scaledFace, 0)

	// Draw the number with rotation so it aligns radially
	op := &text.DrawOptions{}

	// First translate to center of text, rotate, then translate to position
	// The text should be rotated so it points outward from the center
	op.GeoM.Translate(-textWidth/2, -textHeight/2)
	op.GeoM.Rotate(angle + math.Pi/2) // Rotate to align with slot (add 90 degrees so text reads outward)
	op.GeoM.Translate(x, y)

	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, numStr, scaledFace, op)
}

// drawCenter draws the center hub of the wheel
func (w *Wheel) drawCenter(screen *ebiten.Image) {
	centerR := w.Radius * CenterRatio

	// Outer gold ring
	drawFilledCircle(screen, w.CenterX, w.CenterY, centerR, ColorGold)

	// Inner decorative area
	drawFilledCircle(screen, w.CenterX, w.CenterY, centerR*0.8, ColorWoodDark)

	// Center jewel/boss
	drawFilledCircle(screen, w.CenterX, w.CenterY, centerR*0.3, ColorChrome)
	drawFilledCircle(screen, w.CenterX, w.CenterY, centerR*0.2, ColorGold)

	// Decorative spokes
	numSpokes := 8
	for i := 0; i < numSpokes; i++ {
		angle := float64(i) * math.Pi / 4
		x1 := w.CenterX + centerR*0.35*math.Cos(angle)
		y1 := w.CenterY + centerR*0.35*math.Sin(angle)
		x2 := w.CenterX + centerR*0.75*math.Cos(angle)
		y2 := w.CenterY + centerR*0.75*math.Sin(angle)
		vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 3, ColorGold, false)
	}
}

// Helper drawing functions

func drawFilledCircle(screen *ebiten.Image, cx, cy, radius float64, clr color.Color) {
	vector.DrawFilledCircle(screen, float32(cx), float32(cy), float32(radius), clr, false)
}

func drawRing(screen *ebiten.Image, cx, cy, outerR, innerR float64, clr color.Color) {
	// Draw as a thick stroked circle
	midR := (outerR + innerR) / 2
	thickness := outerR - innerR
	vector.StrokeCircle(screen, float32(cx), float32(cy), float32(midR), float32(thickness), clr, false)
}

func drawWedge(screen *ebiten.Image, cx, cy, innerR, outerR, startAngle, endAngle float64, clr color.Color) {
	// Draw wedge using small triangles
	steps := 20
	angleStep := (endAngle - startAngle) / float64(steps)

	for i := 0; i < steps; i++ {
		a1 := startAngle + float64(i)*angleStep
		a2 := startAngle + float64(i+1)*angleStep

		// Inner and outer points
		ix1, iy1 := cx+innerR*math.Cos(a1), cy+innerR*math.Sin(a1)
		ix2, iy2 := cx+innerR*math.Cos(a2), cy+innerR*math.Sin(a2)
		ox1, oy1 := cx+outerR*math.Cos(a1), cy+outerR*math.Sin(a1)
		ox2, oy2 := cx+outerR*math.Cos(a2), cy+outerR*math.Sin(a2)

		// Draw two triangles to form a quad
		drawTriangle(screen, ix1, iy1, ox1, oy1, ox2, oy2, clr)
		drawTriangle(screen, ix1, iy1, ox2, oy2, ix2, iy2, clr)
	}
}

func drawTriangle(screen *ebiten.Image, x1, y1, x2, y2, x3, y3 float64, clr color.Color) {
	path := &vector.Path{}
	path.MoveTo(float32(x1), float32(y1))
	path.LineTo(float32(x2), float32(y2))
	path.LineTo(float32(x3), float32(y3))
	path.Close()

	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	r, g, b, a := clr.RGBA()
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = float32(r) / 0xffff
		vs[i].ColorG = float32(g) / 0xffff
		vs[i].ColorB = float32(b) / 0xffff
		vs[i].ColorA = float32(a) / 0xffff
	}

	op := &ebiten.DrawTrianglesOptions{}
	op.FillRule = ebiten.FillRuleNonZero
	screen.DrawTriangles(vs, is, whiteImage, op)
}

func drawDiamond(screen *ebiten.Image, cx, cy, size, angle float64, clr color.Color) {
	// Diamond shape rotated to point outward
	pts := []struct{ x, y float64 }{
		{0, -size * 1.5},
		{size, 0},
		{0, size * 1.5},
		{-size, 0},
	}

	// Rotate points
	for i := range pts {
		rx := pts[i].x*math.Cos(angle) - pts[i].y*math.Sin(angle)
		ry := pts[i].x*math.Sin(angle) + pts[i].y*math.Cos(angle)
		pts[i].x = cx + rx
		pts[i].y = cy + ry
	}

	// Draw as two triangles
	drawTriangle(screen, pts[0].x, pts[0].y, pts[1].x, pts[1].y, pts[2].x, pts[2].y, clr)
	drawTriangle(screen, pts[0].x, pts[0].y, pts[2].x, pts[2].y, pts[3].x, pts[3].y, clr)
}

// whiteImage is a 1x1 white image used as a base for drawing
var whiteImage = func() *ebiten.Image {
	img := ebiten.NewImage(3, 3)
	img.Fill(color.White)
	return img
}()

// GetNumberColor returns the color for a given number string
func GetNumberColor(numStr string) color.RGBA {
	if numStr == "0" || numStr == "00" {
		return ColorGreen
	}
	if RedNumbers[numStr] {
		return ColorRed
	}
	return ColorBlack
}

// IsRed returns true if the number is red
func IsRed(numStr string) bool {
	return RedNumbers[numStr]
}

// IsEven returns true if the number is even (0 and 00 are neither)
func IsEven(numStr string) bool {
	if numStr == "0" || numStr == "00" {
		return false
	}
	n := 0
	for _, c := range numStr {
		n = n*10 + int(c-'0')
	}
	return n%2 == 0
}

// IsLow returns true if the number is in the low range (1-18)
func IsLow(numStr string) bool {
	if numStr == "0" || numStr == "00" {
		return false
	}
	n := 0
	for _, c := range numStr {
		n = n*10 + int(c-'0')
	}
	return n >= 1 && n <= 18
}
