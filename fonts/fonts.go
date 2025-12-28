// Package fonts provides premium font management for the roulette wheel application.
// It embeds a high-quality Inter Bold font and provides pre-rendered faces at multiple sizes.
package fonts

import (
	"bytes"
	_ "embed"
	"log"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed Inter-Bold.ttf
var interBoldTTF []byte

// FontSize represents standard font sizes used throughout the application
type FontSize int

const (
	SizeSmall      FontSize = 14  // Controls, small labels
	SizeBody       FontSize = 18  // Default body text
	SizeMedium     FontSize = 24  // Sub-headers, emphasized text
	SizeLarge      FontSize = 36  // Section headers
	SizeXLarge     FontSize = 48  // Large numbers, titles
	SizeHuge       FontSize = 72  // Large display numbers
	SizeWinSmall   FontSize = 96  // Winning animation smaller
	SizeWinLarge   FontSize = 144 // Winning animation full size
)

// Manager holds pre-rendered font faces at various sizes
type Manager struct {
	source *text.GoTextFaceSource
	faces  map[FontSize]*text.GoTextFace
}

// NewManager creates a font manager with pre-rendered faces
func NewManager() *Manager {
	source, err := text.NewGoTextFaceSource(bytes.NewReader(interBoldTTF))
	if err != nil {
		log.Printf("Failed to load Inter Bold font: %v", err)
		return nil
	}

	m := &Manager{
		source: source,
		faces:  make(map[FontSize]*text.GoTextFace),
	}

	// Pre-render all standard sizes
	sizes := []FontSize{
		SizeSmall,
		SizeBody,
		SizeMedium,
		SizeLarge,
		SizeXLarge,
		SizeHuge,
		SizeWinSmall,
		SizeWinLarge,
	}

	for _, size := range sizes {
		m.faces[size] = &text.GoTextFace{
			Source: source,
			Size:   float64(size),
		}
	}

	return m
}

// Face returns the pre-rendered font face for the given size
func (m *Manager) Face(size FontSize) *text.GoTextFace {
	if m == nil || m.faces == nil {
		return nil
	}
	return m.faces[size]
}

// FaceForSize returns a font face for an arbitrary size.
// For best performance, prefer using standard FontSize constants.
func (m *Manager) FaceForSize(size float64) *text.GoTextFace {
	if m == nil || m.source == nil {
		return nil
	}
	// Check if we have a pre-rendered face close to this size
	for fontSize, face := range m.faces {
		if float64(fontSize) == size {
			return face
		}
	}
	// Create a new face for this size
	return &text.GoTextFace{
		Source: m.source,
		Size:   size,
	}
}

// Source returns the font source for creating custom-sized faces
func (m *Manager) Source() *text.GoTextFaceSource {
	if m == nil {
		return nil
	}
	return m.source
}
