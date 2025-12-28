<objective>
Implement realistic roulette wheel audio using procedural synthesis. Add a continuous rolling/rumbling sound while the ball orbits, enhance the bounce sound with metallic harmonics, and synchronize audio with ball physics.
</objective>

<context>
This is a Go/Ebitengine roulette wheel game. The current audio system in `./audio/audio.go` uses procedural synthesis for tick, bounce, settle, and chime sounds.

Research completed in `./research/audio-options.md` recommends:
- Filtered noise synthesis for rolling sound (no external audio files)
- Enhanced bounce with metallic ring harmonics
- Real-time volume modulation based on ball angular speed

Files to examine:
- `./audio/audio.go` - Current audio implementation
- `./ball/ball.go` - Ball phases and callbacks
- `./main.go` - Game loop where audio is wired up
</context>

<requirements>
1. **Add Rolling Sound** (continuous while ball orbits):
   - Generate using filtered white noise (low-pass filter for brown-ish rumble)
   - Create a short loopable buffer (~500ms) using `audio.NewInfiniteLoop()`
   - Add `StartRolling()`, `StopRolling()`, `UpdateRollingVolume()` methods
   - Volume should scale with ball's angular speed (faster = louder)
   - Start when ball enters PhaseOrbiting, stop when it settles

2. **Enhance Bounce Sound**:
   - Add metallic ring harmonics (2400Hz, 3600Hz, 4800Hz) with longer decay
   - Keep the impact thud but layer metallic ping on top
   - Should sound like a small metal ball hitting metal slots

3. **Wire Up Synchronization** in main.go:
   - Call `StartRolling()` when ball starts orbiting
   - Call `UpdateRollingVolume(speedRatio)` each frame during orbit
   - Call `StopRolling()` when ball enters bouncing or settled phase

4. **Handle Edge Cases**:
   - Respect mute setting for rolling sound
   - Clean up rolling player on game reset
   - Prevent multiple rolling sounds from overlapping
</requirements>

<implementation>
Reference the code examples in `./research/audio-options.md` section 4.2 for:
- `generateRollingLoop()` - Filtered noise with fade in/out for seamless loop
- `StartRolling()`, `StopRolling()`, `UpdateRollingVolume()` - Player management
- Enhanced `generateBounce()` - Impact + metallic ring + noise

Key technical details:
- Filter coefficient ~0.95 for dark, woody rumble
- Use `audio.NewInfiniteLoop()` for seamless rolling loop
- `player.SetVolume()` for real-time volume control
- Map `ball.AngularSpeed / InitialAngularSpeed` to volume ratio
</implementation>

<verification>
Build and run with `go build && ./roulette-wheel`:
1. Spin the wheel and listen for continuous rolling sound during orbit phase
2. Rolling sound should get quieter as ball slows down
3. Rolling should stop when ball starts bouncing
4. Bounce sounds should have a metallic "ping" quality
5. Mute (M key) should silence all sounds including rolling
6. Reset (R key) should properly stop rolling sound
7. Multiple spins should not cause overlapping rolling sounds
</verification>

<success_criteria>
- Continuous rolling/rumbling heard while ball orbits outer track
- Rolling volume decreases as ball slows
- Bounce sounds have metallic character
- No audio glitches, clicks, or overlapping sounds
- All existing sounds (tick, settle, chime) still work
- Mute and reset functions work correctly
</success_criteria>
