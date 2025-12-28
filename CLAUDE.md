# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
make build      # Build to dist/roulette-wheel
make run        # Run with go run .
make clean      # Remove dist directory
go build -o roulette-wheel .  # Direct build
```

## Architecture

Go application using Ebitengine for 2D graphics and audio. The game loop follows Ebitengine's `Update()` (game logic at 60 TPS) and `Draw()` (rendering) pattern.

### Package Structure

- **main.go** - Game state machine coordinating wheel, ball, stats, and audio. Manages spin lifecycle: `isSpinning` -> `ballSettled` -> `resultDeclared`.
- **wheel/** - Wheel rendering with 38-slot American layout. Uses `NumberSequence` for slot order and `RedNumbers` map for colors. Rotates via `Rotation` angle in radians.
- **ball/** - Physics simulation with phase-based state machine: `PhaseIdle` -> `PhaseOrbiting` -> `PhaseDropping` -> `PhaseBouncing` -> `PhaseSettled`. Uses crypto/rand for unpredictable results.
- **stats/** - Statistics tracking (history, hot/cold numbers, percentages) and right-side panel rendering.
- **audio/** - Programmatically generated sounds (tick, bounce, settle, chime, rolling loop) using raw PCM samples at 44.1kHz stereo.

### Key Constants

Wheel dimensions use ratios relative to `Radius` (e.g., `BallTrackOuterRatio = 1.0`, `SlotOuterRatio = 0.88`). Ball physics constants like `InitialAngularSpeed` and `DropThreshold` control spin behavior.

### Callbacks

Ball emits events (`OnTick`, `OnBounce`, `OnSettle`) that main.go wires to audio playback.
