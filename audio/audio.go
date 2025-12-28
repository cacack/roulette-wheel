// Package audio handles sound effect management for the roulette wheel.
package audio

import (
	"bytes"
	"encoding/binary"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// Audio constants
const (
	SampleRate = 44100
)

// Audio represents the audio manager
type Audio struct {
	context     *audio.Context
	muted       bool
	mu          sync.Mutex

	// Pre-generated sound buffers
	tickSound   []byte
	bounceSound []byte
	settleSound []byte
	chimeSound  []byte

	// Players for concurrent playback
	tickPlayer   *audio.Player
	bouncePlayer *audio.Player
	settlePlayer *audio.Player
	chimePlayer  *audio.Player
}

// New creates a new audio manager
func New() *Audio {
	ctx := audio.NewContext(SampleRate)

	a := &Audio{
		context: ctx,
	}

	// Generate sounds in the background
	go a.generateSounds()

	return a
}

// generateSounds creates all sound effects programmatically
func (a *Audio) generateSounds() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.tickSound = generateTick()
	a.bounceSound = generateBounce()
	a.settleSound = generateSettle()
	a.chimeSound = generateChime()
}

// PlayTick plays the wheel tick sound
func (a *Audio) PlayTick() {
	a.playSound(a.tickSound)
}

// PlayBounce plays the ball bounce sound
func (a *Audio) PlayBounce() {
	a.playSound(a.bounceSound)
}

// PlaySettle plays the ball settling sound
func (a *Audio) PlaySettle() {
	a.playSound(a.settleSound)
}

// PlayChime plays the win chime sound
func (a *Audio) PlayChime() {
	a.playSound(a.chimeSound)
}

// playSound plays a sound buffer
func (a *Audio) playSound(soundData []byte) {
	if a.muted || len(soundData) == 0 {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.context == nil {
		return
	}

	// Create a new player for each sound to allow overlapping
	player, err := a.context.NewPlayer(bytes.NewReader(soundData))
	if err != nil {
		return
	}

	player.Play()
}

// ToggleMute toggles the mute state
func (a *Audio) ToggleMute() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.muted = !a.muted
}

// IsMuted returns the current mute state
func (a *Audio) IsMuted() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.muted
}

// SetMuted sets the mute state
func (a *Audio) SetMuted(muted bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.muted = muted
}

// Sound generation functions

// generateTick creates a short clicking sound
func generateTick() []byte {
	duration := 0.02 // 20ms
	samples := int(SampleRate * duration)

	buf := make([]byte, samples*4) // Stereo 16-bit

	for i := 0; i < samples; i++ {
		t := float64(i) / SampleRate

		// Short click with fast decay
		envelope := math.Exp(-t * 200)

		// Mix of frequencies for a wooden click
		sample := math.Sin(2*math.Pi*1500*t)*0.6 +
			math.Sin(2*math.Pi*3000*t)*0.3 +
			math.Sin(2*math.Pi*600*t)*0.1

		sample *= envelope * 0.3

		// Convert to 16-bit
		val := int16(sample * 32767)

		// Stereo output
		binary.LittleEndian.PutUint16(buf[i*4:], uint16(val))
		binary.LittleEndian.PutUint16(buf[i*4+2:], uint16(val))
	}

	return buf
}

// generateBounce creates a ball bounce sound
func generateBounce() []byte {
	duration := 0.1 // 100ms
	samples := int(SampleRate * duration)

	buf := make([]byte, samples*4)

	for i := 0; i < samples; i++ {
		t := float64(i) / SampleRate

		// Impact with quick decay
		envelope := math.Exp(-t * 40)

		// Low thud with some higher harmonics
		sample := math.Sin(2*math.Pi*200*t)*0.5 +
			math.Sin(2*math.Pi*400*t)*0.3 +
			math.Sin(2*math.Pi*800*t)*0.15 +
			(randomNoise()*0.05)*envelope // Add some noise for realism

		sample *= envelope * 0.4

		val := int16(sample * 32767)

		binary.LittleEndian.PutUint16(buf[i*4:], uint16(val))
		binary.LittleEndian.PutUint16(buf[i*4+2:], uint16(val))
	}

	return buf
}

// generateSettle creates the final settling sound
func generateSettle() []byte {
	duration := 0.15 // 150ms
	samples := int(SampleRate * duration)

	buf := make([]byte, samples*4)

	for i := 0; i < samples; i++ {
		t := float64(i) / SampleRate

		// Soft thud with slow decay
		envelope := math.Exp(-t * 25)

		sample := math.Sin(2*math.Pi*150*t)*0.6 +
			math.Sin(2*math.Pi*300*t)*0.3 +
			math.Sin(2*math.Pi*100*t)*0.1

		sample *= envelope * 0.35

		val := int16(sample * 32767)

		binary.LittleEndian.PutUint16(buf[i*4:], uint16(val))
		binary.LittleEndian.PutUint16(buf[i*4+2:], uint16(val))
	}

	return buf
}

// generateChime creates a win announcement chime
func generateChime() []byte {
	duration := 0.8 // 800ms
	samples := int(SampleRate * duration)

	buf := make([]byte, samples*4)

	// Musical notes for a pleasant chime (C major arpeggio)
	notes := []float64{523.25, 659.25, 783.99} // C5, E5, G5

	for i := 0; i < samples; i++ {
		t := float64(i) / SampleRate

		// Envelope with attack and decay
		attack := 0.02
		var envelope float64
		if t < attack {
			envelope = t / attack
		} else {
			envelope = math.Exp(-(t - attack) * 3)
		}

		// Sum of harmonious notes
		sample := 0.0
		for _, freq := range notes {
			sample += math.Sin(2*math.Pi*freq*t) * 0.25
			sample += math.Sin(2*math.Pi*freq*2*t) * 0.1 // Harmonic
		}

		sample *= envelope * 0.3

		val := int16(sample * 32767)

		binary.LittleEndian.PutUint16(buf[i*4:], uint16(val))
		binary.LittleEndian.PutUint16(buf[i*4+2:], uint16(val))
	}

	return buf
}

// Simple deterministic noise for audio (not crypto-random, just for audio texture)
var noiseState uint32 = 12345

func randomNoise() float64 {
	noiseState = noiseState*1103515245 + 12345
	return float64(noiseState&0x7FFFFFFF)/float64(0x7FFFFFFF)*2 - 1
}
