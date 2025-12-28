// Package stats handles statistics tracking and display for the roulette wheel.
package stats

import (
	"fmt"
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
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

	// Display
	PanelX      float64
	PanelY      float64
	PanelWidth  float64
	PanelHeight float64
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
func New(panelX, panelY, panelWidth, panelHeight float64) *Stats {
	return &Stats{
		History:     make([]string, 0, 20),
		Counts:      make(map[string]int),
		MaxHistory:  15,
		PanelX:      panelX,
		PanelY:      panelY,
		PanelWidth:  panelWidth,
		PanelHeight: panelHeight,
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

// UpdatePosition updates the panel position (for window resize)
func (s *Stats) UpdatePosition(panelX, panelY, panelWidth, panelHeight float64) {
	s.PanelX = panelX
	s.PanelY = panelY
	s.PanelWidth = panelWidth
	s.PanelHeight = panelHeight
}

// Draw renders the statistics panel
func (s *Stats) Draw(screen *ebiten.Image, fontFace *text.GoTextFace) {
	// Draw panel background
	vector.DrawFilledRect(screen, float32(s.PanelX), float32(s.PanelY),
		float32(s.PanelWidth), float32(s.PanelHeight), ColorPanel, false)

	// Draw border
	vector.StrokeRect(screen, float32(s.PanelX), float32(s.PanelY),
		float32(s.PanelWidth), float32(s.PanelHeight), 2, ColorGold, false)

	y := s.PanelY + 30
	padding := 20.0
	lineHeight := 35.0

	// Title
	s.drawText(screen, fontFace, "STATISTICS", s.PanelX+padding, y, ColorGold, 1.5)
	y += lineHeight * 1.5

	// Total spins
	s.drawText(screen, fontFace, fmt.Sprintf("Total Spins: %d", s.TotalSpins), s.PanelX+padding, y, ColorText, 1.0)
	y += lineHeight * 1.2

	// Last Number (large display)
	s.drawText(screen, fontFace, "Last Number:", s.PanelX+padding, y, ColorText, 1.0)
	y += lineHeight * 1.2
	if len(s.History) > 0 {
		lastNum := s.History[len(s.History)-1]
		s.drawLastNumber(screen, fontFace, lastNum, s.PanelX+padding, y)
	}
	y += lineHeight * 2.8

	// History
	s.drawText(screen, fontFace, "History:", s.PanelX+padding, y, ColorText, 1.0)
	y += lineHeight * 0.8
	s.drawHistory(screen, fontFace, y)
	y += lineHeight * 1.5

	// Hot Numbers
	s.drawText(screen, fontFace, "Hot Numbers:", s.PanelX+padding, y, ColorHot, 1.0)
	y += lineHeight * 0.8
	hotNums := s.GetHotNumbers(5)
	s.drawNumberList(screen, fontFace, hotNums, s.PanelX+padding, y)
	y += lineHeight * 1.2

	// Cold Numbers
	s.drawText(screen, fontFace, "Cold Numbers:", s.PanelX+padding, y, ColorCold, 1.0)
	y += lineHeight * 0.8
	coldNums := s.GetColdNumbers(5)
	s.drawNumberList(screen, fontFace, coldNums, s.PanelX+padding, y)
	y += lineHeight * 1.5

	// Percentage bars
	s.drawPercentageBar(screen, fontFace, "Red/Black", s.PanelX+padding, y, s.PanelWidth-padding*2,
		s.RedCount, s.BlackCount, ColorRed, ColorBlack)
	y += lineHeight * 1.5

	s.drawPercentageBar(screen, fontFace, "Even/Odd", s.PanelX+padding, y, s.PanelWidth-padding*2,
		s.EvenCount, s.OddCount, color.RGBA{30, 60, 120, 255}, color.RGBA{100, 150, 220, 255})
	y += lineHeight * 1.5

	s.drawPercentageBar(screen, fontFace, "Low/High", s.PanelX+padding, y, s.PanelWidth-padding*2,
		s.LowCount, s.HighCount, color.RGBA{30, 60, 120, 255}, color.RGBA{100, 150, 220, 255})
	y += lineHeight * 1.5

	// Green (0/00) count
	if s.TotalSpins > 0 {
		greenPct := float64(s.GreenCount) / float64(s.TotalSpins) * 100
		s.drawText(screen, fontFace, fmt.Sprintf("Green (0/00): %d (%.1f%%)", s.GreenCount, greenPct),
			s.PanelX+padding, y, ColorGreen, 1.0)
	}
}

func (s *Stats) drawText(screen *ebiten.Image, fontFace *text.GoTextFace, str string, x, y float64, clr color.Color, scale float64) {
	if fontFace == nil {
		// Fallback: draw a simple indicator
		return
	}

	op := &text.DrawOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, str, fontFace, op)
}

func (s *Stats) drawLastNumber(screen *ebiten.Image, fontFace *text.GoTextFace, number string, x, y float64) {
	// Draw large colored chip for last number
	chipSize := 40.0
	centerX := x + chipSize
	centerY := y + chipSize/2

	// Chip background
	chipColor := getNumberColor(number)
	vector.DrawFilledCircle(screen, float32(centerX), float32(centerY), float32(chipSize), chipColor, false)
	vector.StrokeCircle(screen, float32(centerX), float32(centerY), float32(chipSize), 3, ColorGold, false)

	// Number text centered in circle
	textScale := 2.0
	// Better centering: offset from center based on text length
	textOffsetX := -11.0 * float64(len(number))
	textOffsetY := -14.0
	s.drawText(screen, fontFace, number, centerX+textOffsetX, centerY+textOffsetY, ColorText, textScale)
}

func (s *Stats) drawHistory(screen *ebiten.Image, fontFace *text.GoTextFace, y float64) {
	chipSize := 18.0
	spacing := chipSize*2 + 5
	x := s.PanelX + 20

	// Draw history from newest to oldest
	for i := len(s.History) - 1; i >= 0; i-- {
		num := s.History[i]
		chipColor := getNumberColor(num)

		centerX := x + chipSize
		centerY := y + chipSize

		// Draw chip
		vector.DrawFilledCircle(screen, float32(centerX), float32(centerY), float32(chipSize), chipColor, false)
		vector.StrokeCircle(screen, float32(centerX), float32(centerY), float32(chipSize), 1, ColorGold, false)

		// Draw number text centered in circle
		textScale := 0.8
		// Better centering: offset from center based on text length
		textOffsetX := -5.0 * float64(len(num))
		textOffsetY := -6.0
		s.drawText(screen, fontFace, num, centerX+textOffsetX, centerY+textOffsetY, ColorText, textScale)

		x += spacing
		if x > s.PanelX+s.PanelWidth-40 {
			break // Don't overflow panel
		}
	}
}

func (s *Stats) drawNumberList(screen *ebiten.Image, fontFace *text.GoTextFace, numbers []string, x, y float64) {
	chipSize := 16.0
	spacing := chipSize*2 + 8

	for _, num := range numbers {
		chipColor := getNumberColor(num)

		centerX := x + chipSize
		centerY := y + chipSize

		// Draw chip
		vector.DrawFilledCircle(screen, float32(centerX), float32(centerY), float32(chipSize), chipColor, false)
		vector.StrokeCircle(screen, float32(centerX), float32(centerY), float32(chipSize), 1, ColorGold, false)

		// Draw number text centered in circle
		textScale := 0.7
		// Better centering: offset from center based on text length
		textOffsetX := -4.0 * float64(len(num))
		textOffsetY := -5.0
		s.drawText(screen, fontFace, num, centerX+textOffsetX, centerY+textOffsetY, ColorText, textScale)

		x += spacing
	}
}

func (s *Stats) drawPercentageBar(screen *ebiten.Image, fontFace *text.GoTextFace, label string, x, y, width float64,
	count1, count2 int, color1, color2 color.Color) {

	// Label
	s.drawText(screen, fontFace, label, x, y, ColorText, 0.9)
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

		// Percentage labels
		pct1Text := fmt.Sprintf("%.0f%%", pct1*100)
		pct2Text := fmt.Sprintf("%.0f%%", (1-pct1)*100)
		s.drawText(screen, fontFace, pct1Text, x+5, y+2, ColorText, 0.8)
		s.drawText(screen, fontFace, pct2Text, x+width-35, y+2, ColorText, 0.8)
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
