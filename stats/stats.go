// Package stats handles statistics tracking and display for the roulette wheel.
package stats

import (
	"fmt"
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"roulette-wheel/fonts"
)

// Stats tracks roulette statistics
type Stats struct {
	History      []string       // Last N results
	Counts       map[string]int // Count per number
	TotalSpins   int
	MaxHistory   int

	// Running totals
	RedCount   int
	BlackCount int
	GreenCount int
	EvenCount  int
	OddCount   int
	LowCount   int  // 1-18
	HighCount  int  // 19-36

	// Right panel display (stats)
	PanelX      float64
	PanelY      float64
	PanelWidth  float64
	PanelHeight float64

	// Left panel display (history)
	HistoryPanelX      float64
	HistoryPanelY      float64
	HistoryPanelWidth  float64
	HistoryPanelHeight float64
}

// Number color mapping
var redNumbers = map[string]bool{
	"1": true, "3": true, "5": true, "7": true, "9": true,
	"12": true, "14": true, "16": true, "18": true, "19": true,
	"21": true, "23": true, "25": true, "27": true, "30": true,
	"32": true, "34": true, "36": true,
}

// Color constants for display
var (
	ColorRed      = color.RGBA{185, 30, 30, 255}
	ColorBlack    = color.RGBA{25, 25, 25, 255}
	ColorGreen    = color.RGBA{0, 128, 0, 255}
	ColorGold     = color.RGBA{218, 165, 32, 255}
	ColorText     = color.RGBA{255, 255, 255, 255}
	ColorPanel    = color.RGBA{20, 40, 20, 240}
	ColorBarBg    = color.RGBA{50, 50, 50, 255}
	ColorHot      = color.RGBA{255, 100, 50, 255}
	ColorCold     = color.RGBA{50, 150, 255, 255}
)

// New creates a new stats tracker
func New(panelX, panelY, panelWidth, panelHeight, histPanelX, histPanelY, histPanelWidth, histPanelHeight float64) *Stats {
	return &Stats{
		History:            make([]string, 0, 20),
		Counts:             make(map[string]int),
		MaxHistory:         15,
		PanelX:             panelX,
		PanelY:             panelY,
		PanelWidth:         panelWidth,
		PanelHeight:        panelHeight,
		HistoryPanelX:      histPanelX,
		HistoryPanelY:      histPanelY,
		HistoryPanelWidth:  histPanelWidth,
		HistoryPanelHeight: histPanelHeight,
	}
}

// RecordResult adds a new result to the statistics
func (s *Stats) RecordResult(number string) {
	s.TotalSpins++

	// Add to history
	s.History = append(s.History, number)
	if len(s.History) > s.MaxHistory {
		s.History = s.History[1:]
	}

	// Update counts
	s.Counts[number]++

	// Update category counts
	if number == "0" || number == "00" {
		s.GreenCount++
	} else if redNumbers[number] {
		s.RedCount++
	} else {
		s.BlackCount++
	}

	// Parse number for even/odd and high/low
	if number != "0" && number != "00" {
		n := parseNumber(number)
		if n%2 == 0 {
			s.EvenCount++
		} else {
			s.OddCount++
		}
		if n >= 1 && n <= 18 {
			s.LowCount++
		} else if n >= 19 && n <= 36 {
			s.HighCount++
		}
	}
}

// Reset clears all statistics
func (s *Stats) Reset() {
	s.History = s.History[:0]
	s.Counts = make(map[string]int)
	s.TotalSpins = 0
	s.RedCount = 0
	s.BlackCount = 0
	s.GreenCount = 0
	s.EvenCount = 0
	s.OddCount = 0
	s.LowCount = 0
	s.HighCount = 0
}

// GetHotNumbers returns the most frequently hit numbers
func (s *Stats) GetHotNumbers(n int) []string {
	return s.getTopNumbers(n, true)
}

// GetColdNumbers returns the least frequently hit numbers (including those never hit)
func (s *Stats) GetColdNumbers(n int) []string {
	// All possible roulette numbers
	allNumbers := []string{
		"0", "00", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10",
		"11", "12", "13", "14", "15", "16", "17", "18", "19", "20",
		"21", "22", "23", "24", "25", "26", "27", "28", "29", "30",
		"31", "32", "33", "34", "35", "36",
	}

	type numCount struct {
		num   string
		count int
	}

	var counts []numCount
	for _, num := range allNumbers {
		counts = append(counts, numCount{num, s.Counts[num]})
	}

	// Sort by count ascending (least frequent first), with secondary sort by number for stability
	sort.Slice(counts, func(i, j int) bool {
		if counts[i].count != counts[j].count {
			return counts[i].count < counts[j].count
		}
		// When counts are equal, sort by number for stability
		return counts[i].num < counts[j].num
	})

	result := make([]string, 0, n)
	for i := 0; i < n && i < len(counts); i++ {
		result = append(result, counts[i].num)
	}
	return result
}

func (s *Stats) getTopNumbers(n int, hot bool) []string {
	type numCount struct {
		num   string
		count int
	}

	var counts []numCount
	for num, count := range s.Counts {
		counts = append(counts, numCount{num, count})
	}

	// Sort with secondary key (number) to ensure stable ordering
	sort.Slice(counts, func(i, j int) bool {
		if counts[i].count != counts[j].count {
			if hot {
				return counts[i].count > counts[j].count
			}
			return counts[i].count < counts[j].count
		}
		// When counts are equal, sort by number for stability
		return counts[i].num < counts[j].num
	})

	result := make([]string, 0, n)
	for i := 0; i < n && i < len(counts); i++ {
		result = append(result, counts[i].num)
	}
	return result
}

// UpdatePosition updates both panel positions (for window resize)
func (s *Stats) UpdatePosition(panelX, panelY, panelWidth, panelHeight, histPanelX, histPanelY, histPanelWidth, histPanelHeight float64) {
	s.PanelX = panelX
	s.PanelY = panelY
	s.PanelWidth = panelWidth
	s.PanelHeight = panelHeight
	s.HistoryPanelX = histPanelX
	s.HistoryPanelY = histPanelY
	s.HistoryPanelWidth = histPanelWidth
	s.HistoryPanelHeight = histPanelHeight
}

// Draw renders both panels (history on left, stats on right)
func (s *Stats) Draw(screen *ebiten.Image, fontMgr *fonts.Manager) {
	// Draw left history panel
	s.DrawHistoryPanel(screen, fontMgr)

	// Draw right stats panel
	s.DrawStatsPanel(screen, fontMgr)
}

// DrawHistoryPanel renders the vertical history panel on the left side
func (s *Stats) DrawHistoryPanel(screen *ebiten.Image, fontMgr *fonts.Manager) {
	// Draw panel background
	vector.DrawFilledRect(screen, float32(s.HistoryPanelX), float32(s.HistoryPanelY),
		float32(s.HistoryPanelWidth), float32(s.HistoryPanelHeight), ColorPanel, false)

	// Draw border
	vector.StrokeRect(screen, float32(s.HistoryPanelX), float32(s.HistoryPanelY),
		float32(s.HistoryPanelWidth), float32(s.HistoryPanelHeight), 2, ColorGold, false)

	padding := 10.0
	y := s.HistoryPanelY + 25
	centerX := s.HistoryPanelX + s.HistoryPanelWidth/2

	// "LAST" section title with proper font sizing
	titleFace := fontMgr.Face(fonts.SizeMedium) // 24pt for section title
	s.drawTextWithFace(screen, titleFace, "LAST", s.HistoryPanelX+padding+15, y, ColorGold)
	y += 35

	if len(s.History) > 0 {
		lastNum := s.History[len(s.History)-1]
		lastColor := getNumberColor(lastNum)
		lastChipSize := 38.0

		// Draw large chip for last number
		vector.DrawFilledCircle(screen, float32(centerX), float32(y+lastChipSize), float32(lastChipSize), lastColor, false)
		vector.StrokeCircle(screen, float32(centerX), float32(y+lastChipSize), float32(lastChipSize), 2, ColorGold, false)

		// Draw number text centered in circle - use properly sized font
		chipFace := fontMgr.Face(fonts.SizeLarge) // 36pt for large chip number
		textWidth, textHeight := text.Measure(lastNum, chipFace, 0)
		chipTextX := centerX - textWidth/2
		chipTextY := y + lastChipSize - textHeight/2
		s.drawTextWithFace(screen, chipFace, lastNum, chipTextX, chipTextY, ColorText)

		y += lastChipSize*2 + 15
	} else {
		y += 30
	}

	// Separator line
	vector.StrokeLine(screen, float32(s.HistoryPanelX+10), float32(y),
		float32(s.HistoryPanelX+s.HistoryPanelWidth-10), float32(y), 1, ColorGold, false)
	y += 15

	// "HISTORY" title with proper font sizing
	historyTitleFace := fontMgr.Face(fonts.SizeBody) // 18pt for smaller title
	s.drawTextWithFace(screen, historyTitleFace, "HISTORY", s.HistoryPanelX+padding+10, y, ColorGold)
	y += 30

	// Draw history vertically (newest at top, skip first since it's shown in LAST)
	chipSize := 20.0
	spacing := chipSize*2 + 6

	// Use small font for history chip numbers
	historyFace := fontMgr.Face(fonts.SizeSmall) // 14pt for history chips

	startIdx := len(s.History) - 2 // Skip the most recent (shown in LAST)
	for i := startIdx; i >= 0; i-- {
		// Stop if we would overflow the panel
		if y+chipSize*2 > s.HistoryPanelY+s.HistoryPanelHeight-20 {
			break
		}

		num := s.History[i]
		chipColor := getNumberColor(num)

		// Draw chip centered in panel
		vector.DrawFilledCircle(screen, float32(centerX), float32(y+chipSize), float32(chipSize), chipColor, false)
		vector.StrokeCircle(screen, float32(centerX), float32(y+chipSize), float32(chipSize), 1.5, ColorGold, false)

		// Draw number text centered in circle - properly measured
		textWidth, textHeight := text.Measure(num, historyFace, 0)
		textX := centerX - textWidth/2
		textY := y + chipSize - textHeight/2
		s.drawTextWithFace(screen, historyFace, num, textX, textY, ColorText)

		y += spacing
	}
}

// DrawStatsPanel renders the statistics panel on the right side
func (s *Stats) DrawStatsPanel(screen *ebiten.Image, fontMgr *fonts.Manager) {
	// Draw panel background
	vector.DrawFilledRect(screen, float32(s.PanelX), float32(s.PanelY),
		float32(s.PanelWidth), float32(s.PanelHeight), ColorPanel, false)

	// Draw border
	vector.StrokeRect(screen, float32(s.PanelX), float32(s.PanelY),
		float32(s.PanelWidth), float32(s.PanelHeight), 2, ColorGold, false)

	y := s.PanelY + 30
	padding := 20.0
	lineHeight := 35.0

	// Get font faces for different sizes
	titleFace := fontMgr.Face(fonts.SizeLarge)   // 36pt for main title
	headerFace := fontMgr.Face(fonts.SizeBody)   // 18pt for section headers
	bodyFace := fontMgr.Face(fonts.SizeBody)     // 18pt for body text
	smallFace := fontMgr.Face(fonts.SizeSmall)   // 14pt for bar labels

	// Title - "STATISTICS" with premium look
	s.drawTextWithFace(screen, titleFace, "STATISTICS", s.PanelX+padding, y, ColorGold)
	y += lineHeight * 1.8

	// Total spins
	s.drawTextWithFace(screen, bodyFace, fmt.Sprintf("Total Spins: %d", s.TotalSpins), s.PanelX+padding, y, ColorText)
	y += lineHeight * 1.5

	// Hot Numbers
	s.drawTextWithFace(screen, headerFace, "Hot Numbers:", s.PanelX+padding, y, ColorHot)
	y += lineHeight * 0.8
	hotNums := s.GetHotNumbers(5)
	s.drawNumberListWithFonts(screen, fontMgr, hotNums, s.PanelX+padding, y)
	y += lineHeight * 1.2

	// Cold Numbers
	s.drawTextWithFace(screen, headerFace, "Cold Numbers:", s.PanelX+padding, y, ColorCold)
	y += lineHeight * 0.8
	coldNums := s.GetColdNumbers(5)
	s.drawNumberListWithFonts(screen, fontMgr, coldNums, s.PanelX+padding, y)
	y += lineHeight * 1.5

	// Percentage bars
	s.drawPercentageBarWithFonts(screen, bodyFace, smallFace, "Red/Black", s.PanelX+padding, y, s.PanelWidth-padding*2,
		s.RedCount, s.BlackCount, ColorRed, ColorBlack)
	y += lineHeight * 1.5

	s.drawPercentageBarWithFonts(screen, bodyFace, smallFace, "Even/Odd", s.PanelX+padding, y, s.PanelWidth-padding*2,
		s.EvenCount, s.OddCount, color.RGBA{30, 60, 120, 255}, color.RGBA{100, 150, 220, 255})
	y += lineHeight * 1.5

	s.drawPercentageBarWithFonts(screen, bodyFace, smallFace, "Low/High", s.PanelX+padding, y, s.PanelWidth-padding*2,
		s.LowCount, s.HighCount, color.RGBA{30, 60, 120, 255}, color.RGBA{100, 150, 220, 255})
	y += lineHeight * 1.5

	// Green (0/00) count
	if s.TotalSpins > 0 {
		greenPct := float64(s.GreenCount) / float64(s.TotalSpins) * 100
		s.drawTextWithFace(screen, bodyFace, fmt.Sprintf("Green (0/00): %d (%.1f%%)", s.GreenCount, greenPct),
			s.PanelX+padding, y, ColorGreen)
	}
}

// drawTextWithFace draws text using a specific font face without scaling
func (s *Stats) drawTextWithFace(screen *ebiten.Image, face *text.GoTextFace, str string, x, y float64, clr color.Color) {
	if face == nil {
		return
	}

	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, str, face, op)
}

// drawNumberListWithFonts draws chip numbers using proper font sizing
func (s *Stats) drawNumberListWithFonts(screen *ebiten.Image, fontMgr *fonts.Manager, numbers []string, x, y float64) {
	chipSize := 16.0
	spacing := chipSize*2 + 8

	// Use small font for chip numbers
	face := fontMgr.Face(fonts.SizeSmall) // 14pt

	for _, num := range numbers {
		chipColor := getNumberColor(num)

		centerX := x + chipSize
		centerY := y + chipSize

		// Draw chip
		vector.DrawFilledCircle(screen, float32(centerX), float32(centerY), float32(chipSize), chipColor, false)
		vector.StrokeCircle(screen, float32(centerX), float32(centerY), float32(chipSize), 1, ColorGold, false)

		// Draw number text centered in circle - properly measured
		textWidth, textHeight := text.Measure(num, face, 0)
		textX := centerX - textWidth/2
		textY := centerY - textHeight/2
		s.drawTextWithFace(screen, face, num, textX, textY, ColorText)

		x += spacing
	}
}

// drawPercentageBarWithFonts draws a percentage bar with proper font sizing
func (s *Stats) drawPercentageBarWithFonts(screen *ebiten.Image, labelFace, pctFace *text.GoTextFace, label string, x, y, width float64,
	count1, count2 int, color1, color2 color.Color) {

	// Label
	s.drawTextWithFace(screen, labelFace, label, x, y, ColorText)
	y += 22

	barHeight := 20.0
	total := count1 + count2

	// Background
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(width), float32(barHeight), ColorBarBg, false)

	if total > 0 {
		// First color portion
		pct1 := float64(count1) / float64(total)
		width1 := width * pct1
		vector.DrawFilledRect(screen, float32(x), float32(y), float32(width1), float32(barHeight), color1, false)

		// Second color portion
		vector.DrawFilledRect(screen, float32(x+width1), float32(y), float32(width-width1), float32(barHeight), color2, false)

		// Percentage labels - properly positioned
		pct1Text := fmt.Sprintf("%.0f%%", pct1*100)
		pct2Text := fmt.Sprintf("%.0f%%", (1-pct1)*100)

		// Left percentage
		s.drawTextWithFace(screen, pctFace, pct1Text, x+5, y+3, ColorText)

		// Right percentage - measure to right-align
		pct2Width, _ := text.Measure(pct2Text, pctFace, 0)
		s.drawTextWithFace(screen, pctFace, pct2Text, x+width-pct2Width-5, y+3, ColorText)
	}

	// Border
	vector.StrokeRect(screen, float32(x), float32(y), float32(width), float32(barHeight), 1, ColorGold, false)
}

// GetLastNumber returns the most recent result
func (s *Stats) GetLastNumber() string {
	if len(s.History) == 0 {
		return ""
	}
	return s.History[len(s.History)-1]
}

// Helper functions

func getNumberColor(num string) color.RGBA {
	if num == "0" || num == "00" {
		return ColorGreen
	}
	if redNumbers[num] {
		return ColorRed
	}
	return ColorBlack
}

func parseNumber(s string) int {
	n := 0
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n
}
