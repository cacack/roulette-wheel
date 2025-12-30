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

// Wood grain color palette for realistic mahogany/walnut finish
var (
	// Outer rim wood tones (lighter mahogany)
	woodRimLight    = color.RGBA{165, 110, 60, 255}
	woodRimMid      = color.RGBA{139, 90, 43, 255}
	woodRimDark     = color.RGBA{115, 70, 35, 255}
	woodRimDeep     = color.RGBA{95, 55, 25, 255}

	// Ball track wood tones (darker walnut)
	woodTrackLight  = color.RGBA{85, 55, 35, 255}
	woodTrackMid    = color.RGBA{60, 40, 25, 255}
	woodTrackDark   = color.RGBA{45, 30, 18, 255}
	woodTrackDeep   = color.RGBA{35, 22, 12, 255}

	// Center/slot area wood tones (rich dark wood)
	woodCenterLight = color.RGBA{120, 80, 45, 255}
	woodCenterMid   = color.RGBA{101, 67, 33, 255}
	woodCenterDark  = color.RGBA{80, 52, 25, 255}
	woodCenterDeep  = color.RGBA{60, 38, 18, 255}

	// Lacquer highlight colors
	lacquerHighlight = color.RGBA{255, 255, 240, 40}
	lacquerSubtle    = color.RGBA{255, 255, 230, 20}

	// Chrome/brushed metal color palette (for axle, dividers, rings)
	chromeDark      = color.RGBA{90, 95, 100, 255}     // Deep shadow
	chromeMid       = color.RGBA{140, 145, 150, 255}   // Mid-tone brushed steel
	chromeLight     = color.RGBA{190, 195, 200, 255}   // Bright chrome
	chromeBright    = color.RGBA{230, 235, 240, 255}   // Highlight
	chromeSpecular  = color.RGBA{255, 255, 255, 200}   // Specular highlight

	// Gold color palette (for polished gold surfaces)
	goldDark        = color.RGBA{160, 120, 20, 255}    // Deep gold shadow
	goldMid         = color.RGBA{200, 155, 30, 255}    // Mid gold
	goldLight       = color.RGBA{230, 190, 50, 255}    // Bright gold
	goldBright      = color.RGBA{255, 220, 100, 255}   // Gold highlight
	goldSpecular    = color.RGBA{255, 250, 200, 180}   // Gold specular

	// Shadow and lighting colors for 3D depth effects
	// Light source is upper-left (around -135 degrees / -3π/4)
	shadowDark      = color.RGBA{0, 0, 0, 50}         // Deep shadow for recessed areas
	shadowMid       = color.RGBA{0, 0, 0, 30}         // Medium shadow
	shadowLight     = color.RGBA{0, 0, 0, 18}         // Subtle shadow
	shadowVeryLight = color.RGBA{0, 0, 0, 10}         // Very subtle ambient occlusion
	ambientLight    = color.RGBA{255, 255, 240, 25}   // Warm ambient light from upper-left
	ambientSubtle   = color.RGBA{255, 255, 240, 12}   // Very subtle ambient
)

// Wheel dimensions as ratios of the wheel radius
const (
	BallTrackOuterRatio = 1.0
	BallTrackInnerRatio = 0.92
	SlotOuterRatio      = 0.88
	SlotInnerRatio      = 0.65
	CenterRatio         = 0.30
	DeflectorRadiusRatio = 0.90  // Between ball track and slot area
	NumDeflectors       = 8
	DeflectorHitRadius  = 0.06  // Hit zone radius ratio (slightly larger than visual)
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
	FontSource    *text.GoTextFaceSource // Font source for creating sized faces

	// Cached rendering for performance
	cachedImage  *ebiten.Image // Pre-rendered wheel at rotation=0
	cachedRadius float64       // Radius when cache was created
}

// New creates a new wheel at the given position and size
func New(centerX, centerY, radius float64) *Wheel {
	return &Wheel{
		CenterX: centerX,
		CenterY: centerY,
		Radius:  radius,
	}
}

// Update updates the wheel state (fixed timestep - no dt parameter)
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

// ensureCache creates or updates the cached wheel image if needed
func (w *Wheel) ensureCache() {
	if w.cachedImage != nil && w.cachedRadius == w.Radius {
		return // Cache is valid
	}

	// Create offscreen image large enough for the wheel
	// Add padding for any anti-aliasing artifacts
	size := int(w.Radius*2) + 20
	w.cachedImage = ebiten.NewImage(size, size)
	w.cachedRadius = w.Radius

	// Draw wheel centered in the cache at rotation=0
	cacheCenter := float64(size) / 2

	// Wood color palettes for different wheel areas
	rimPalette := []color.RGBA{woodRimLight, woodRimMid, woodRimDark, woodRimDeep}
	trackPalette := []color.RGBA{woodTrackLight, woodTrackMid, woodTrackDark, woodTrackDeep}
	centerPalette := []color.RGBA{woodCenterLight, woodCenterMid, woodCenterDark, woodCenterDeep}

	// Draw ball track outer wooden rim with grain texture
	outerRimRadius := w.Radius * BallTrackOuterRatio
	innerRimRadius := w.Radius * BallTrackInnerRatio
	drawWoodGrain(w.cachedImage, cacheCenter, cacheCenter, outerRimRadius, innerRimRadius, rimPalette)
	drawLacquerHighlight(w.cachedImage, cacheCenter, cacheCenter, outerRimRadius, innerRimRadius)

	// Draw ball track inner surface (where ball rolls) with darker grain
	trackOuterRadius := innerRimRadius
	trackInnerRadius := w.Radius * SlotOuterRatio
	drawWoodGrain(w.cachedImage, cacheCenter, cacheCenter, trackOuterRadius, trackInnerRadius, trackPalette)
	drawLacquerHighlight(w.cachedImage, cacheCenter, cacheCenter, trackOuterRadius, trackInnerRadius)

	// Draw deflectors (at rotation=0)
	w.drawDeflectorsToCache(w.cachedImage, cacheCenter)

	// Draw slot area background with wood grain (dark rich wood around slots)
	slotOuterRadius := w.Radius * SlotOuterRatio
	slotInnerRadius := w.Radius * SlotInnerRatio
	drawWoodGrain(w.cachedImage, cacheCenter, cacheCenter, slotOuterRadius, slotInnerRadius, centerPalette)

	// Draw shadow where ball track rim meets slot area (rim casts shadow on slots)
	drawBallTrackRimShadow(w.cachedImage, cacheCenter, cacheCenter, slotOuterRadius)

	// Draw individual slots (at rotation=0) - includes pocket depth shadows
	w.drawSlotsToCache(w.cachedImage, cacheCenter)

	// Draw center hub shadow BEFORE drawing the hub (shadow appears behind it)
	centerR := w.Radius * CenterRatio
	drawCenterHubShadow(w.cachedImage, cacheCenter, cacheCenter, centerR)

	// Draw center cone/hub
	w.drawCenterToCache(w.cachedImage, cacheCenter)

	// Draw beveled chrome ring separating slots from center
	chromeRingOuter := w.Radius * SlotInnerRatio
	chromeRingInner := chromeRingOuter - 6
	chromePalette := []color.RGBA{chromeDark, chromeMid, chromeLight, chromeBright}
	drawBrushedMetalRing(w.cachedImage, cacheCenter, cacheCenter, chromeRingOuter, chromeRingInner, chromePalette)

	// Add beveled edge to chrome ring
	bevelWidth := 1.5
	drawBeveledRing(w.cachedImage, cacheCenter, cacheCenter, chromeRingOuter, chromeRingInner, chromeLight, bevelWidth)

	// Add specular highlight to chrome ring
	drawMetalSpecular(w.cachedImage, cacheCenter, cacheCenter, chromeRingOuter, chromeRingInner, chromeSpecular, -math.Pi*0.7, -math.Pi*0.35)

	// Apply overall directional lighting as final pass
	// This creates cohesive lighting across all wheel elements
	drawOverallWheelLighting(w.cachedImage, cacheCenter, cacheCenter, w.Radius)
}

// InvalidateCache forces the wheel to be re-rendered (call after resize)
func (w *Wheel) InvalidateCache() {
	w.cachedImage = nil
}

// Draw renders the wheel using the cached image
func (w *Wheel) Draw(screen *ebiten.Image) {
	w.ensureCache()

	// Calculate where to draw the cached image
	cacheSize := float64(w.cachedImage.Bounds().Dx())
	cacheCenter := cacheSize / 2

	// Draw cached wheel with rotation
	op := &ebiten.DrawImageOptions{}
	// Translate so rotation center is at origin
	op.GeoM.Translate(-cacheCenter, -cacheCenter)
	// Apply rotation
	op.GeoM.Rotate(w.Rotation)
	// Translate to screen position
	op.GeoM.Translate(w.CenterX, w.CenterY)

	screen.DrawImage(w.cachedImage, op)
}

// drawDeflectorsToCache draws deflectors to the cached image at rotation=0
func (w *Wheel) drawDeflectorsToCache(img *ebiten.Image, center float64) {
	deflectorAngle := 2 * math.Pi / float64(NumDeflectors)
	deflectorRadius := w.Radius * DeflectorRadiusRatio

	for i := 0; i < NumDeflectors; i++ {
		angle := float64(i) * deflectorAngle // No w.Rotation - cache is at rotation=0
		x := center + deflectorRadius*math.Cos(angle)
		y := center + deflectorRadius*math.Sin(angle)

		// Larger diamond size - approximately 2x ball diameter
		size := w.Radius * 0.045

		// Alternate orientation: even = vertical (points radially), odd = horizontal (points tangentially)
		diamondAngle := angle
		if i%2 == 1 {
			diamondAngle += math.Pi / 2 // Rotate 90 degrees for horizontal orientation
		}

		// Draw polished chrome deflector diamond with enhanced 3D effect
		drawPolishedChromeDiamond(img, x, y, size, diamondAngle)
	}
}

// drawPolishedChromeDiamond draws a deflector as a polished chrome diamond with gradient and highlights
func drawPolishedChromeDiamond(screen *ebiten.Image, cx, cy, size, angle float64) {
	// Shadow layer - offset and darker
	shadowOffset := size * 0.15
	shadowX := cx + shadowOffset*0.7
	shadowY := cy + shadowOffset
	shadowColor := color.RGBA{40, 45, 50, 180}
	drawDiamond(screen, shadowX, shadowY, size*1.05, angle, shadowColor)

	// Base chrome layer - dark edge
	drawDiamond(screen, cx, cy, size, angle, chromeDark)

	// Mid chrome layer - smaller, lighter
	drawDiamond(screen, cx, cy, size*0.85, angle, chromeMid)

	// Main chrome surface - convex center
	drawDiamond(screen, cx, cy, size*0.7, angle, chromeLight)

	// Bright core - center highlight
	drawDiamond(screen, cx, cy, size*0.5, angle, chromeBright)

	// Specular highlight - small bright spot offset toward light source (upper-left)
	highlightOffsetX := -size * 0.2
	highlightOffsetY := -size * 0.25
	drawDiamond(screen, cx+highlightOffsetX, cy+highlightOffsetY, size*0.25, angle, chromeSpecular)

	// Secondary reflection - subtle lower-right ambient
	ambientColor := color.RGBA{200, 205, 210, 80}
	drawDiamond(screen, cx+size*0.15, cy+size*0.2, size*0.15, angle, ambientColor)
}

// drawSlotsToCache draws all 38 numbered slots to the cached image at rotation=0
func (w *Wheel) drawSlotsToCache(img *ebiten.Image, center float64) {
	outerR := w.Radius * SlotOuterRatio
	innerR := w.Radius * SlotInnerRatio

	for i, numStr := range NumberSequence {
		startAngle := float64(i)*SlotAngle - SlotAngle/2 // No w.Rotation
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
		drawWedge(img, center, center, innerR, outerR, startAngle, endAngle, slotColor)

		// Add pocket depth shadow to create recessed appearance
		drawSlotPocketShadow(img, center, center, innerR, outerR, startAngle, endAngle)

		// Draw slot dividers as beveled metal strips
		dividerAngle := float64(i)*SlotAngle + SlotAngle/2 // No w.Rotation
		drawBeveledSlotDivider(img, center, center, innerR, outerR, dividerAngle)

		// Draw number
		w.drawNumberToCache(img, center, i, numStr)
	}
}

// drawNumberToCache draws a number on a slot to the cached image at rotation=0
func (w *Wheel) drawNumberToCache(img *ebiten.Image, center float64, slotIndex int, numStr string) {
	if w.FontSource == nil {
		return
	}

	// Position at middle of slot
	midRadius := w.Radius * (SlotOuterRatio + SlotInnerRatio) / 2
	angle := float64(slotIndex) * SlotAngle // No w.Rotation

	x := center + midRadius*math.Cos(angle)
	y := center + midRadius*math.Sin(angle)

	// Calculate font size based on wheel radius (scale with wheel size)
	// Use a size that fits well in the slot width
	fontSize := w.Radius * 0.05
	if fontSize < 10 {
		fontSize = 10
	}
	if fontSize > 28 {
		fontSize = 28
	}

	// Create a properly sized font face for the number
	face := &text.GoTextFace{
		Source: w.FontSource,
		Size:   fontSize,
	}

	// Measure text to center it properly
	textWidth, textHeight := text.Measure(numStr, face, 0)

	// Draw shadow for better readability
	shadowOp := &text.DrawOptions{}
	shadowOffset := fontSize * 0.05
	shadowOp.GeoM.Translate(-textWidth/2+shadowOffset, -textHeight/2+shadowOffset)
	shadowOp.GeoM.Rotate(angle + math.Pi/2) // Rotate to align with slot
	shadowOp.GeoM.Translate(x, y)
	shadowOp.ColorScale.ScaleWithColor(color.RGBA{0, 0, 0, 120})
	text.Draw(img, numStr, face, shadowOp)

	// Draw the number with rotation so it aligns radially
	op := &text.DrawOptions{}
	op.GeoM.Translate(-textWidth/2, -textHeight/2)
	op.GeoM.Rotate(angle + math.Pi/2) // Rotate to align with slot (add 90 degrees so text reads outward)
	op.GeoM.Translate(x, y)

	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(img, numStr, face, op)
}

// drawCenterToCache draws the center hub to the cached image
func (w *Wheel) drawCenterToCache(img *ebiten.Image, center float64) {
	centerR := w.Radius * CenterRatio
	centerPalette := []color.RGBA{woodCenterLight, woodCenterMid, woodCenterDark, woodCenterDeep}

	// Outer polished gold ring with gradient
	drawPolishedGoldRing(img, center, center, centerR, centerR*0.82)

	// Add beveled edge to gold ring for 3D depth
	bevelWidth := centerR * 0.03
	drawBeveledRing(img, center, center, centerR, centerR*0.82, goldMid, bevelWidth)

	// Inner decorative area with wood grain
	innerWoodOuter := centerR * 0.8
	innerWoodInner := centerR * 0.35
	drawWoodGrain(img, center, center, innerWoodOuter, innerWoodInner, centerPalette)
	drawLacquerHighlight(img, center, center, innerWoodOuter, innerWoodInner)

	// Center jewel/boss - polished chrome outer ring
	// Use simple smooth gradient instead of brushed metal for this small area
	chromeOuterR := centerR * 0.32
	chromeInnerR := centerR * 0.22
	drawSmoothChromeRing(img, center, center, chromeOuterR, chromeInnerR)

	// Add chrome ring highlight
	drawMetalSpecular(img, center, center, chromeOuterR, chromeInnerR, chromeSpecular, -math.Pi*0.65, -math.Pi*0.35)

	// Center polished gold boss (convex dome)
	drawPolishedGoldCircle(img, center, center, centerR*0.2)

	// Decorative spokes (polished gold inlay on wood)
	numSpokes := 8
	for i := 0; i < numSpokes; i++ {
		angle := float64(i) * math.Pi / 4
		drawPolishedGoldSpoke(img, center, center, centerR*0.35, centerR*0.75, angle)
	}
}

// drawPolishedGoldSpoke draws a decorative gold spoke with 3D beveled effect
func drawPolishedGoldSpoke(screen *ebiten.Image, cx, cy, innerR, outerR, angle float64) {
	perpAngle := angle + math.Pi/2
	spokeWidth := 1.5

	// Light edge
	lightOffset := spokeWidth * 0.6
	lx1 := cx + innerR*math.Cos(angle) + lightOffset*math.Cos(perpAngle)
	ly1 := cy + innerR*math.Sin(angle) + lightOffset*math.Sin(perpAngle)
	lx2 := cx + outerR*math.Cos(angle) + lightOffset*math.Cos(perpAngle)
	ly2 := cy + outerR*math.Sin(angle) + lightOffset*math.Sin(perpAngle)
	vector.StrokeLine(screen, float32(lx1), float32(ly1), float32(lx2), float32(ly2), 1.0, goldBright, false)

	// Main gold strip
	x1 := cx + innerR*math.Cos(angle)
	y1 := cy + innerR*math.Sin(angle)
	x2 := cx + outerR*math.Cos(angle)
	y2 := cy + outerR*math.Sin(angle)
	vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 3.0, goldLight, false)

	// Shadow edge
	shadowOffset := -spokeWidth * 0.6
	sx1 := cx + innerR*math.Cos(angle) + shadowOffset*math.Cos(perpAngle)
	sy1 := cy + innerR*math.Sin(angle) + shadowOffset*math.Sin(perpAngle)
	sx2 := cx + outerR*math.Cos(angle) + shadowOffset*math.Cos(perpAngle)
	sy2 := cy + outerR*math.Sin(angle) + shadowOffset*math.Sin(perpAngle)
	vector.StrokeLine(screen, float32(sx1), float32(sy1), float32(sx2), float32(sy2), 1.0, goldDark, false)
}

// Helper drawing functions

func drawFilledCircle(screen *ebiten.Image, cx, cy, radius float64, clr color.Color) {
	vector.DrawFilledCircle(screen, float32(cx), float32(cy), float32(radius), clr, false)
}

// interpolateColor blends two colors based on t (0.0 = c1, 1.0 = c2)
func interpolateColor(c1, c2 color.RGBA, t float64) color.RGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return color.RGBA{
		R: uint8(float64(c1.R) + t*(float64(c2.R)-float64(c1.R))),
		G: uint8(float64(c1.G) + t*(float64(c2.G)-float64(c1.G))),
		B: uint8(float64(c1.B) + t*(float64(c2.B)-float64(c1.B))),
		A: 255,
	}
}

// getWoodGrainColor returns a wood grain color based on ring index using a repeating pattern
// The pattern creates light->dark->light grain bands like real lathe-turned wood
func getWoodGrainColor(ringIndex int, colors []color.RGBA, depthFactor float64) color.RGBA {
	// Create a sine-wave pattern for natural grain variation
	// Use multiple frequencies to create more organic, less uniform grain
	primary := math.Sin(float64(ringIndex)*0.8) * 0.5 + 0.5   // Main grain pattern
	secondary := math.Sin(float64(ringIndex)*1.7) * 0.15      // Secondary variation
	tertiary := math.Sin(float64(ringIndex)*0.3) * 0.1        // Slow undulation
	t := primary + secondary + tertiary

	// Clamp to 0-1 range
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	// Map to color palette (4 colors: light, mid, dark, deep)
	if len(colors) < 4 {
		return colors[0]
	}

	// Blend through the palette based on combined wave position
	var baseColor color.RGBA
	if t < 0.33 {
		baseColor = interpolateColor(colors[0], colors[1], t*3)
	} else if t < 0.66 {
		baseColor = interpolateColor(colors[1], colors[2], (t-0.33)*3)
	} else {
		baseColor = interpolateColor(colors[2], colors[3], (t-0.66)*3)
	}

	// Apply depth darkening - wood appears slightly darker at edges due to curvature
	// depthFactor: 0 = outer edge (slightly darker), 1 = inner edge
	edgeDarken := 1.0 - depthFactor*0.08 - (1.0-depthFactor)*0.05
	return color.RGBA{
		R: uint8(float64(baseColor.R) * edgeDarken),
		G: uint8(float64(baseColor.G) * edgeDarken),
		B: uint8(float64(baseColor.B) * edgeDarken),
		A: 255,
	}
}

// drawWoodGrain draws concentric wood grain rings simulating lathe-turned hardwood
func drawWoodGrain(screen *ebiten.Image, cx, cy, outerR, innerR float64, palette []color.RGBA) {
	// Calculate ring width - thinner rings = finer grain detail
	ringWidth := 2.0 // pixels per grain ring
	numRings := int((outerR - innerR) / ringWidth)
	if numRings < 1 {
		numRings = 1
	}

	// Draw rings from outside in
	totalWidth := outerR - innerR
	for i := 0; i < numRings; i++ {
		radius := outerR - float64(i)*ringWidth
		if radius < innerR {
			break
		}
		// Calculate depth factor (0 at outer edge, 1 at inner edge)
		depthFactor := float64(i) * ringWidth / totalWidth
		clr := getWoodGrainColor(i, palette, depthFactor)
		drawRing(screen, cx, cy, radius, radius-ringWidth, clr)
	}
}

// drawLacquerHighlight draws a subtle highlight arc on a circular surface
// to simulate light reflecting off polished lacquered wood
func drawLacquerHighlight(screen *ebiten.Image, cx, cy, outerR, innerR float64) {
	// Draw highlight arcs on upper-left portion (simulating overhead light source)
	// Real lacquered wood has multiple highlight bands from the curved surface

	midR := (outerR + innerR) / 2
	highlightWidth := (outerR - innerR) * 0.25

	// Primary bright highlight - the main reflection point
	startAngle := -math.Pi * 0.75
	endAngle := -math.Pi * 0.45
	drawHighlightArc(screen, cx, cy, midR, highlightWidth, startAngle, endAngle, lacquerHighlight)

	// Outer edge highlight - fainter, wider arc on the outer part of the surface
	outerHighlightR := outerR - (outerR-innerR)*0.15
	drawHighlightArc(screen, cx, cy, outerHighlightR, highlightWidth*0.6, -math.Pi*0.85, -math.Pi*0.55, lacquerSubtle)

	// Inner edge highlight - fainter on the inner curve
	innerHighlightR := innerR + (outerR-innerR)*0.15
	drawHighlightArc(screen, cx, cy, innerHighlightR, highlightWidth*0.5, -math.Pi*0.65, -math.Pi*0.35, lacquerSubtle)

	// Subtle ambient reflection on opposite side (from environment)
	startAngle2 := math.Pi * 0.2
	endAngle2 := math.Pi * 0.4
	ambientHighlight := color.RGBA{255, 255, 245, 12}
	drawHighlightArc(screen, cx, cy, midR, highlightWidth*0.8, startAngle2, endAngle2, ambientHighlight)
}

// drawHighlightArc draws a curved highlight arc segment
func drawHighlightArc(screen *ebiten.Image, cx, cy, radius, width, startAngle, endAngle float64, clr color.RGBA) {
	steps := 30
	angleStep := (endAngle - startAngle) / float64(steps)

	innerR := radius - width/2
	outerR := radius + width/2

	for i := 0; i < steps; i++ {
		a1 := startAngle + float64(i)*angleStep
		a2 := startAngle + float64(i+1)*angleStep

		// Fade alpha at edges for soft highlight
		edgeFade := 1.0
		distFromCenter := math.Abs(float64(i) - float64(steps)/2) / (float64(steps) / 2)
		edgeFade = 1.0 - distFromCenter*distFromCenter // Quadratic falloff

		fadeClr := color.RGBA{clr.R, clr.G, clr.B, uint8(float64(clr.A) * edgeFade)}

		// Draw arc segment
		ix1, iy1 := cx+innerR*math.Cos(a1), cy+innerR*math.Sin(a1)
		ix2, iy2 := cx+innerR*math.Cos(a2), cy+innerR*math.Sin(a2)
		ox1, oy1 := cx+outerR*math.Cos(a1), cy+outerR*math.Sin(a1)
		ox2, oy2 := cx+outerR*math.Cos(a2), cy+outerR*math.Sin(a2)

		drawTriangle(screen, ix1, iy1, ox1, oy1, ox2, oy2, fadeClr)
		drawTriangle(screen, ix1, iy1, ox2, oy2, ix2, iy2, fadeClr)
	}
}

func drawRing(screen *ebiten.Image, cx, cy, outerR, innerR float64, clr color.Color) {
	// Draw as a thick stroked circle
	midR := (outerR + innerR) / 2
	thickness := outerR - innerR
	vector.StrokeCircle(screen, float32(cx), float32(cy), float32(midR), float32(thickness), clr, false)
}

// drawSmoothChromeRing draws a chrome ring with a clean, smooth gradient
// Uses fewer rings with smoother color transitions for small areas
func drawSmoothChromeRing(screen *ebiten.Image, cx, cy, outerR, innerR float64) {
	ringWidth := outerR - innerR
	// Use only 3-4 bands for a clean, smooth look
	numBands := 4

	for i := 0; i < numBands; i++ {
		t := float64(i) / float64(numBands)
		tNext := float64(i+1) / float64(numBands)
		radius := outerR - t*ringWidth
		nextRadius := outerR - tNext*ringWidth
		if nextRadius < innerR {
			nextRadius = innerR
		}

		// Simple gradient: outer dark, inner bright (convex appearance)
		var clr color.RGBA
		switch i {
		case 0:
			clr = chromeDark
		case 1:
			clr = chromeMid
		case 2:
			clr = chromeLight
		case 3:
			clr = chromeBright
		default:
			clr = chromeLight
		}
		drawRing(screen, cx, cy, radius, nextRadius, clr)
	}
}

// drawBrushedMetalRing draws a ring with radial brushed metal gradient effect
// The gradient goes from dark edges to bright center, simulating concave polished metal
func drawBrushedMetalRing(screen *ebiten.Image, cx, cy, outerR, innerR float64, palette []color.RGBA) {
	ringWidth := outerR - innerR
	numRings := int(ringWidth / 1.5) // Finer rings for smoother gradient
	if numRings < 3 {
		numRings = 3
	}

	for i := 0; i < numRings; i++ {
		t := float64(i) / float64(numRings-1)
		radius := outerR - t*ringWidth
		nextRadius := radius - ringWidth/float64(numRings)
		if nextRadius < innerR {
			nextRadius = innerR
		}

		// Create concave metal effect: dark at edges, bright in middle
		// Use sin wave for smooth transition
		brightness := math.Sin(t * math.Pi)
		clr := getMetalGradientColor(brightness, palette)
		drawRing(screen, cx, cy, radius, nextRadius, clr)
	}
}

// drawBeveledRing draws a ring with 3D beveled edge effect
// Light edge on top/outer, shadow on bottom/inner for raised appearance
// Uses smooth gradient transitions instead of abrupt highlight/shadow boundaries
func drawBeveledRing(screen *ebiten.Image, cx, cy, outerR, innerR float64, baseColor color.RGBA, bevelWidth float64) {
	highlightColor := color.RGBA{
		R: uint8(min(255, int(baseColor.R)+50)),
		G: uint8(min(255, int(baseColor.G)+50)),
		B: uint8(min(255, int(baseColor.B)+50)),
		A: 255,
	}
	shadowColor := color.RGBA{
		R: uint8(max(0, int(baseColor.R)-60)),
		G: uint8(max(0, int(baseColor.G)-60)),
		B: uint8(max(0, int(baseColor.B)-60)),
		A: 255,
	}

	// Draw outer bevel with smooth gradient around the ring
	highlightOuter := outerR
	highlightInner := outerR - bevelWidth
	drawSmoothBevelArc(screen, cx, cy, highlightOuter, highlightInner, highlightColor, shadowColor, baseColor)

	// Draw main metal surface
	mainOuter := highlightInner
	mainInner := innerR + bevelWidth
	drawRing(screen, cx, cy, mainOuter, mainInner, baseColor)

	// Draw inner bevel (inverted - shadow on top, highlight on bottom for channel effect)
	innerBevelOuter := innerR + bevelWidth
	innerBevelInner := innerR
	drawSmoothBevelArc(screen, cx, cy, innerBevelOuter, innerBevelInner, shadowColor, highlightColor, baseColor)
}

// drawSmoothBevelArc draws a ring with smoothly transitioning highlight/shadow
// Light source assumed from upper-left (-135 degrees / -3π/4)
func drawSmoothBevelArc(screen *ebiten.Image, cx, cy, outerR, innerR float64, highlightColor, shadowColor, baseColor color.RGBA) {
	steps := 60
	angleStep := 2 * math.Pi / float64(steps)
	lightAngle := -math.Pi * 0.75 // Upper-left light source

	for i := 0; i < steps; i++ {
		a1 := float64(i) * angleStep
		a2 := float64(i+1) * angleStep

		// Calculate how much this segment faces the light
		// Use cosine of angle difference for smooth falloff
		midAngle := (a1 + a2) / 2
		lightFacing := math.Cos(midAngle - lightAngle)

		// Interpolate between highlight (lightFacing=1) and shadow (lightFacing=-1)
		// Map from [-1, 1] to [0, 1] for interpolation
		t := (lightFacing + 1) / 2

		// Blend from shadow through base to highlight
		var clr color.RGBA
		if t < 0.5 {
			// Shadow to base (t: 0 -> 0.5 maps to shadow -> base)
			blendT := t * 2
			clr = interpolateColor(shadowColor, baseColor, blendT)
		} else {
			// Base to highlight (t: 0.5 -> 1 maps to base -> highlight)
			blendT := (t - 0.5) * 2
			clr = interpolateColor(baseColor, highlightColor, blendT)
		}

		// Draw this segment
		ix1, iy1 := cx+innerR*math.Cos(a1), cy+innerR*math.Sin(a1)
		ix2, iy2 := cx+innerR*math.Cos(a2), cy+innerR*math.Sin(a2)
		ox1, oy1 := cx+outerR*math.Cos(a1), cy+outerR*math.Sin(a1)
		ox2, oy2 := cx+outerR*math.Cos(a2), cy+outerR*math.Sin(a2)

		drawTriangle(screen, ix1, iy1, ox1, oy1, ox2, oy2, clr)
		drawTriangle(screen, ix1, iy1, ox2, oy2, ix2, iy2, clr)
	}
}

// drawBevelHighlightArc draws a partial arc for bevel lighting effects
func drawBevelHighlightArc(screen *ebiten.Image, cx, cy, outerR, innerR, startAngle, endAngle float64, clr color.RGBA) {
	steps := 40
	angleStep := (endAngle - startAngle) / float64(steps)

	for i := 0; i < steps; i++ {
		a1 := startAngle + float64(i)*angleStep
		a2 := startAngle + float64(i+1)*angleStep

		ix1, iy1 := cx+innerR*math.Cos(a1), cy+innerR*math.Sin(a1)
		ix2, iy2 := cx+innerR*math.Cos(a2), cy+innerR*math.Sin(a2)
		ox1, oy1 := cx+outerR*math.Cos(a1), cy+outerR*math.Sin(a1)
		ox2, oy2 := cx+outerR*math.Cos(a2), cy+outerR*math.Sin(a2)

		drawTriangle(screen, ix1, iy1, ox1, oy1, ox2, oy2, clr)
		drawTriangle(screen, ix1, iy1, ox2, oy2, ix2, iy2, clr)
	}
}

// drawPolishedGoldRing draws a ring with warm polished gold gradient
func drawPolishedGoldRing(screen *ebiten.Image, cx, cy, outerR, innerR float64) {
	ringWidth := outerR - innerR
	numRings := int(ringWidth / 1.5)
	if numRings < 3 {
		numRings = 3
	}

	goldPalette := []color.RGBA{goldDark, goldMid, goldLight, goldBright}

	for i := 0; i < numRings; i++ {
		t := float64(i) / float64(numRings-1)
		radius := outerR - t*ringWidth
		nextRadius := radius - ringWidth/float64(numRings)
		if nextRadius < innerR {
			nextRadius = innerR
		}

		// Convex gold surface: brightest at center bulge
		brightness := math.Sin(t * math.Pi)
		clr := getMetalGradientColor(brightness, goldPalette)
		drawRing(screen, cx, cy, radius, nextRadius, clr)
	}

	// Add specular highlight on upper portion
	drawMetalSpecular(screen, cx, cy, outerR, innerR, goldSpecular, -math.Pi*0.7, -math.Pi*0.4)
}

// drawPolishedChromeCircle draws a filled circle with chrome gradient (convex dome)
// Uses simple 4-band gradient for clean appearance on small surfaces
func drawPolishedChromeCircle(screen *ebiten.Image, cx, cy, radius float64) {
	// Use 4 clean bands for smooth gradient without noise
	bands := []struct {
		ratio float64
		clr   color.RGBA
	}{
		{1.0, chromeDark},
		{0.75, chromeMid},
		{0.5, chromeLight},
		{0.25, chromeBright},
	}

	for i, band := range bands {
		r := radius * band.ratio
		var nextR float64
		if i < len(bands)-1 {
			nextR = radius * bands[i+1].ratio
		} else {
			nextR = 0
		}
		if r > 0 {
			drawFilledCircle(screen, cx, cy, r, band.clr)
		}
		_ = nextR // bands overlap via filled circles
	}

	// Add specular highlight
	highlightOffset := radius * 0.25
	drawFilledCircle(screen, cx-highlightOffset, cy-highlightOffset, radius*0.12, chromeSpecular)
}

// drawPolishedGoldCircle draws a filled circle with gold gradient (convex dome)
func drawPolishedGoldCircle(screen *ebiten.Image, cx, cy, radius float64) {
	goldPalette := []color.RGBA{goldDark, goldMid, goldLight, goldBright}
	numRings := int(radius / 1.5)
	if numRings < 5 {
		numRings = 5
	}

	for i := 0; i < numRings; i++ {
		t := float64(i) / float64(numRings-1)
		r := radius * (1.0 - t)
		nextR := radius * (1.0 - float64(i+1)/float64(numRings-1))
		if nextR < 0 {
			nextR = 0
		}

		// Convex dome: darker at edges, brighter at center
		brightness := t
		clr := getMetalGradientColor(brightness, goldPalette)
		if r > 0 {
			drawRing(screen, cx, cy, r, nextR, clr)
		}
	}

	// Add warm specular highlight
	highlightR := radius * 0.35
	highlightOffset := radius * 0.2
	drawFilledCircle(screen, cx-highlightOffset, cy-highlightOffset, highlightR*0.25, goldSpecular)
}

// getMetalGradientColor returns a color from the metal palette based on brightness (0-1)
func getMetalGradientColor(brightness float64, palette []color.RGBA) color.RGBA {
	if brightness < 0 {
		brightness = 0
	}
	if brightness > 1 {
		brightness = 1
	}

	if len(palette) < 2 {
		return palette[0]
	}

	// Map brightness to palette index
	t := brightness * float64(len(palette)-1)
	idx := int(t)
	if idx >= len(palette)-1 {
		return palette[len(palette)-1]
	}

	frac := t - float64(idx)
	return interpolateColor(palette[idx], palette[idx+1], frac)
}

// drawMetalSpecular draws a specular highlight arc on a metal surface
func drawMetalSpecular(screen *ebiten.Image, cx, cy, outerR, innerR float64, clr color.RGBA, startAngle, endAngle float64) {
	midR := (outerR + innerR) / 2
	width := (outerR - innerR) * 0.3
	drawHighlightArc(screen, cx, cy, midR, width, startAngle, endAngle, clr)
}

// drawBeveledSlotDivider draws a slot divider as a 3D beveled metal strip
func drawBeveledSlotDivider(screen *ebiten.Image, cx, cy, innerR, outerR, angle float64) {
	// Calculate perpendicular offset for 3D effect
	perpAngle := angle + math.Pi/2
	stripWidth := 1.2 // Half-width of the strip

	// Light edge (left side of strip when viewed from center)
	lightOffset := stripWidth * 0.7
	lx1 := cx + innerR*math.Cos(angle) + lightOffset*math.Cos(perpAngle)
	ly1 := cy + innerR*math.Sin(angle) + lightOffset*math.Sin(perpAngle)
	lx2 := cx + outerR*math.Cos(angle) + lightOffset*math.Cos(perpAngle)
	ly2 := cy + outerR*math.Sin(angle) + lightOffset*math.Sin(perpAngle)
	vector.StrokeLine(screen, float32(lx1), float32(ly1), float32(lx2), float32(ly2), 1.0, chromeBright, false)

	// Main chrome strip
	x1 := cx + innerR*math.Cos(angle)
	y1 := cy + innerR*math.Sin(angle)
	x2 := cx + outerR*math.Cos(angle)
	y2 := cy + outerR*math.Sin(angle)
	vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 2.0, chromeLight, false)

	// Shadow edge (right side of strip)
	shadowOffset := -stripWidth * 0.7
	sx1 := cx + innerR*math.Cos(angle) + shadowOffset*math.Cos(perpAngle)
	sy1 := cy + innerR*math.Sin(angle) + shadowOffset*math.Sin(perpAngle)
	sx2 := cx + outerR*math.Cos(angle) + shadowOffset*math.Cos(perpAngle)
	sy2 := cy + outerR*math.Sin(angle) + shadowOffset*math.Sin(perpAngle)
	vector.StrokeLine(screen, float32(sx1), float32(sy1), float32(sx2), float32(sy2), 1.0, chromeDark, false)
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the larger of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
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

// ============================================================================
// Lighting and Shadow Effects
// ============================================================================

// drawShadowRing draws a semi-transparent shadow ring at a boundary
// Creates soft shadow effect for depth at surface transitions
func drawShadowRing(screen *ebiten.Image, cx, cy, outerR, innerR float64, clr color.RGBA) {
	// Draw shadow as a gradient ring that fades from dark to transparent
	ringWidth := outerR - innerR
	numBands := 4
	if ringWidth < 4 {
		numBands = 2
	}

	for i := 0; i < numBands; i++ {
		t := float64(i) / float64(numBands)
		tNext := float64(i+1) / float64(numBands)
		r := outerR - t*ringWidth
		nextR := outerR - tNext*ringWidth
		if nextR < innerR {
			nextR = innerR
		}

		// Fade alpha from center (darker) to edges (lighter)
		// Shadow is darkest in the middle of the band
		midT := (t + tNext) / 2
		alphaFade := 1.0 - math.Abs(midT-0.5)*2 // Peak at center
		bandAlpha := uint8(float64(clr.A) * alphaFade * 0.8)
		bandColor := color.RGBA{clr.R, clr.G, clr.B, bandAlpha}
		drawRing(screen, cx, cy, r, nextR, bandColor)
	}
}

// drawDirectionalShadowArc draws shadow on the side opposite to light source
// Light source is upper-left, so shadows appear on lower-right portions
func drawDirectionalShadowArc(screen *ebiten.Image, cx, cy, outerR, innerR float64, intensity float64) {
	lightAngle := -math.Pi * 0.75 // Upper-left light source
	shadowAngle := lightAngle + math.Pi // Opposite side

	// Shadow arc spans about 120 degrees centered on the shadow direction
	startAngle := shadowAngle - math.Pi/3
	endAngle := shadowAngle + math.Pi/3

	steps := 40
	angleStep := (endAngle - startAngle) / float64(steps)
	midR := (outerR + innerR) / 2
	width := outerR - innerR

	for i := 0; i < steps; i++ {
		a1 := startAngle + float64(i)*angleStep
		a2 := startAngle + float64(i+1)*angleStep

		// Fade at edges of arc
		distFromCenter := math.Abs(float64(i) - float64(steps)/2) / (float64(steps) / 2)
		edgeFade := 1.0 - distFromCenter*distFromCenter

		alpha := uint8(intensity * edgeFade * 255 * 0.15)
		shadowClr := color.RGBA{0, 0, 0, alpha}

		// Draw arc segment
		innerArcR := midR - width/2
		outerArcR := midR + width/2
		ix1, iy1 := cx+innerArcR*math.Cos(a1), cy+innerArcR*math.Sin(a1)
		ix2, iy2 := cx+innerArcR*math.Cos(a2), cy+innerArcR*math.Sin(a2)
		ox1, oy1 := cx+outerArcR*math.Cos(a1), cy+outerArcR*math.Sin(a1)
		ox2, oy2 := cx+outerArcR*math.Cos(a2), cy+outerArcR*math.Sin(a2)

		drawTriangle(screen, ix1, iy1, ox1, oy1, ox2, oy2, shadowClr)
		drawTriangle(screen, ix1, iy1, ox2, oy2, ix2, iy2, shadowClr)
	}
}

// drawDirectionalHighlight draws highlight on the side facing light source
// Light source is upper-left, so highlights appear there
func drawDirectionalHighlight(screen *ebiten.Image, cx, cy, outerR, innerR float64, intensity float64) {
	lightAngle := -math.Pi * 0.75 // Upper-left light source

	// Highlight arc spans about 100 degrees centered on the light direction
	startAngle := lightAngle - math.Pi/3.6
	endAngle := lightAngle + math.Pi/3.6

	steps := 35
	angleStep := (endAngle - startAngle) / float64(steps)
	midR := (outerR + innerR) / 2
	width := (outerR - innerR) * 0.7

	for i := 0; i < steps; i++ {
		a1 := startAngle + float64(i)*angleStep
		a2 := startAngle + float64(i+1)*angleStep

		// Fade at edges of arc
		distFromCenter := math.Abs(float64(i) - float64(steps)/2) / (float64(steps) / 2)
		edgeFade := 1.0 - distFromCenter*distFromCenter

		alpha := uint8(intensity * edgeFade * 255 * 0.12)
		highlightClr := color.RGBA{255, 255, 245, alpha}

		// Draw arc segment
		innerArcR := midR - width/2
		outerArcR := midR + width/2
		ix1, iy1 := cx+innerArcR*math.Cos(a1), cy+innerArcR*math.Sin(a1)
		ix2, iy2 := cx+innerArcR*math.Cos(a2), cy+innerArcR*math.Sin(a2)
		ox1, oy1 := cx+outerArcR*math.Cos(a1), cy+outerArcR*math.Sin(a1)
		ox2, oy2 := cx+outerArcR*math.Cos(a2), cy+outerArcR*math.Sin(a2)

		drawTriangle(screen, ix1, iy1, ox1, oy1, ox2, oy2, highlightClr)
		drawTriangle(screen, ix1, iy1, ox2, oy2, ix2, iy2, highlightClr)
	}
}

// drawCenterHubShadow draws a subtle shadow underneath the center hub
// to make it appear raised above the wheel surface
func drawCenterHubShadow(screen *ebiten.Image, cx, cy, hubRadius float64) {
	// Shadow is offset toward lower-right (opposite of light source)
	shadowOffsetX := hubRadius * 0.08
	shadowOffsetY := hubRadius * 0.1

	// Draw multiple soft shadow rings with decreasing alpha
	// Outer shadow (softest, largest)
	shadowR1 := hubRadius * 1.15
	shadowClr1 := color.RGBA{0, 0, 0, 20}
	drawFilledCircle(screen, cx+shadowOffsetX*1.5, cy+shadowOffsetY*1.5, shadowR1, shadowClr1)

	// Mid shadow
	shadowR2 := hubRadius * 1.08
	shadowClr2 := color.RGBA{0, 0, 0, 25}
	drawFilledCircle(screen, cx+shadowOffsetX, cy+shadowOffsetY, shadowR2, shadowClr2)

	// Inner shadow (darkest, closest to hub edge)
	shadowR3 := hubRadius * 1.02
	shadowClr3 := color.RGBA{0, 0, 0, 18}
	drawFilledCircle(screen, cx+shadowOffsetX*0.5, cy+shadowOffsetY*0.5, shadowR3, shadowClr3)
}

// drawSlotPocketShadow adds subtle shadowing inside a slot to create pocket depth
// Called after the base slot color to darken the outer edge
func drawSlotPocketShadow(screen *ebiten.Image, cx, cy, innerR, outerR, startAngle, endAngle float64) {
	// Draw shadow at the outer edge of the slot (pocket lip shadow)
	shadowDepth := (outerR - innerR) * 0.25
	shadowOuterR := outerR
	shadowInnerR := outerR - shadowDepth

	steps := 12
	angleStep := (endAngle - startAngle) / float64(steps)

	for i := 0; i < steps; i++ {
		a1 := startAngle + float64(i)*angleStep
		a2 := startAngle + float64(i+1)*angleStep

		// Shadow fades from outer edge (dark) inward (transparent)
		// Calculate radial gradient for each triangle

		// Outer edge shadow color (darker)
		shadowAlpha := uint8(35)
		shadowClr := color.RGBA{0, 0, 0, shadowAlpha}

		// Inner edge of shadow (transparent)
		fadeClr := color.RGBA{0, 0, 0, 8}

		// Draw outer shadow band
		ix1, iy1 := cx+shadowInnerR*math.Cos(a1), cy+shadowInnerR*math.Sin(a1)
		ix2, iy2 := cx+shadowInnerR*math.Cos(a2), cy+shadowInnerR*math.Sin(a2)
		ox1, oy1 := cx+shadowOuterR*math.Cos(a1), cy+shadowOuterR*math.Sin(a1)
		ox2, oy2 := cx+shadowOuterR*math.Cos(a2), cy+shadowOuterR*math.Sin(a2)

		// Use average color for the band
		avgAlpha := (shadowAlpha + 8) / 2
		avgClr := color.RGBA{0, 0, 0, avgAlpha}
		_ = fadeClr // we use avgClr for simpler rendering
		_ = shadowClr

		drawTriangle(screen, ix1, iy1, ox1, oy1, ox2, oy2, avgClr)
		drawTriangle(screen, ix1, iy1, ox2, oy2, ix2, iy2, avgClr)
	}

	// Add subtle ambient occlusion at the inner edge too (where slot meets center)
	aoDepth := (outerR - innerR) * 0.15
	aoOuterR := innerR + aoDepth
	aoInnerR := innerR

	for i := 0; i < steps; i++ {
		a1 := startAngle + float64(i)*angleStep
		a2 := startAngle + float64(i+1)*angleStep

		aoClr := color.RGBA{0, 0, 0, 15}

		ix1, iy1 := cx+aoInnerR*math.Cos(a1), cy+aoInnerR*math.Sin(a1)
		ix2, iy2 := cx+aoInnerR*math.Cos(a2), cy+aoInnerR*math.Sin(a2)
		ox1, oy1 := cx+aoOuterR*math.Cos(a1), cy+aoOuterR*math.Sin(a1)
		ox2, oy2 := cx+aoOuterR*math.Cos(a2), cy+aoOuterR*math.Sin(a2)

		drawTriangle(screen, ix1, iy1, ox1, oy1, ox2, oy2, aoClr)
		drawTriangle(screen, ix1, iy1, ox2, oy2, ix2, iy2, aoClr)
	}
}

// drawBallTrackRimShadow draws shadow where ball track rim meets slot area
// Simulates the raised ball track casting shadow on lower slot surface
func drawBallTrackRimShadow(screen *ebiten.Image, cx, cy, slotOuterR float64) {
	// Shadow ring just inside the slot outer edge
	shadowWidth := slotOuterR * 0.04
	shadowOuterR := slotOuterR + 1
	shadowInnerR := slotOuterR - shadowWidth

	// Darker on lower-right (shadow side), lighter on upper-left
	lightAngle := -math.Pi * 0.75

	steps := 60
	angleStep := 2 * math.Pi / float64(steps)

	for i := 0; i < steps; i++ {
		a1 := float64(i) * angleStep
		a2 := float64(i+1) * angleStep
		midAngle := (a1 + a2) / 2

		// Calculate shadow intensity based on angle from light
		// Shadow is strongest opposite to light
		angleDiff := midAngle - lightAngle
		shadowIntensity := (-math.Cos(angleDiff) + 1) / 2 // 0 at light, 1 opposite

		// Base shadow plus directional variation
		baseAlpha := 12.0
		dirAlpha := 20.0 * shadowIntensity
		totalAlpha := uint8(baseAlpha + dirAlpha)

		shadowClr := color.RGBA{0, 0, 0, totalAlpha}

		ix1, iy1 := cx+shadowInnerR*math.Cos(a1), cy+shadowInnerR*math.Sin(a1)
		ix2, iy2 := cx+shadowInnerR*math.Cos(a2), cy+shadowInnerR*math.Sin(a2)
		ox1, oy1 := cx+shadowOuterR*math.Cos(a1), cy+shadowOuterR*math.Sin(a1)
		ox2, oy2 := cx+shadowOuterR*math.Cos(a2), cy+shadowOuterR*math.Sin(a2)

		drawTriangle(screen, ix1, iy1, ox1, oy1, ox2, oy2, shadowClr)
		drawTriangle(screen, ix1, iy1, ox2, oy2, ix2, iy2, shadowClr)
	}
}

// drawOverallWheelLighting applies global directional lighting to the wheel
// Creates cohesive lighting that makes the whole wheel feel 3D
func drawOverallWheelLighting(screen *ebiten.Image, cx, cy, radius float64) {
	// 1. Subtle vignette - very slight darkening at outer edges
	vignetteWidth := radius * 0.08
	vignetteOuterR := radius
	vignetteInnerR := radius - vignetteWidth
	drawDirectionalShadowArc(screen, cx, cy, vignetteOuterR, vignetteInnerR, 0.6)

	// 2. Overall highlight on upper-left portion
	highlightOuterR := radius * 0.95
	highlightInnerR := radius * 0.3
	drawDirectionalHighlight(screen, cx, cy, highlightOuterR, highlightInnerR, 0.5)

	// 3. Subtle shadow on lower-right portion
	shadowOuterR := radius * 0.92
	shadowInnerR := radius * 0.35
	drawDirectionalShadowArc(screen, cx, cy, shadowOuterR, shadowInnerR, 0.4)
}

// DeflectorInfo contains information about a deflector for collision detection
type DeflectorInfo struct {
	Angle       float64 // Angle of deflector center (in wheel-relative coordinates)
	RadiusRatio float64 // Radius ratio where deflector is located
	IsVertical  bool    // True if diamond points radially, false if tangentially
}

// GetDeflectorInfo returns info about all deflectors for collision detection
func GetDeflectorInfo(wheelRotation float64) []DeflectorInfo {
	deflectors := make([]DeflectorInfo, NumDeflectors)
	deflectorAngle := 2 * math.Pi / float64(NumDeflectors)

	for i := 0; i < NumDeflectors; i++ {
		deflectors[i] = DeflectorInfo{
			Angle:       float64(i)*deflectorAngle + wheelRotation,
			RadiusRatio: DeflectorRadiusRatio,
			IsVertical:  i%2 == 0,
		}
	}
	return deflectors
}

// GetDeflectorHitRadius returns the hit zone radius ratio for collision detection
func GetDeflectorHitRadius() float64 {
	return DeflectorHitRadius
}

// GetDeflectorRadiusRatio returns the radius ratio where deflectors are located
func GetDeflectorRadiusRatio() float64 {
	return DeflectorRadiusRatio
}

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
