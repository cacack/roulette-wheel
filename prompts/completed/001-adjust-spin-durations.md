<objective>
Adjust the roulette wheel and ball spin durations so that:
- The main wheel spins for approximately 10 seconds before stopping
- The ball spins for approximately 6 seconds before settling into a slot
</objective>

<context>
This is a roulette wheel game built with Go and Ebitengine (60fps).

Current implementation:
- Ball spin duration: Random 300-600 ticks (5-10 seconds) in `ball/ball.go:105`
- Wheel uses friction-based slowdown with `WheelSpinSpeed = 0.03` in `main.go:27`

The game runs at 60fps, so:
- 10 seconds = 600 ticks
- 6 seconds = 360 ticks
</context>

<requirements>
1. Set ball spin duration to approximately 6 seconds (360 ticks at 60fps)
2. Adjust wheel friction/speed so it spins for approximately 10 seconds
3. Keep the visual feel smooth - the wheel should gradually slow down, not abruptly stop
4. Ball should still settle naturally into a slot after its spin completes
</requirements>

<implementation>
Files to modify:
- `./ball/ball.go` - Adjust SpinDuration calculation around line 105
- `./main.go` or `./wheel/wheel.go` - Adjust wheel friction/speed constants

The ball duration is straightforward - set `SpinDuration` to ~360 ticks with a small random variance for variety.

For the wheel, you may need to adjust `WheelSpinSpeed` and/or the friction coefficient to achieve ~10 second spin time. Test empirically.
</implementation>

<verification>
Run the game with `go build && ./roulette-wheel` and observe:
1. Start a spin and time the wheel - it should take roughly 10 seconds to stop
2. The ball should drop into a slot around 6 seconds after spin starts
3. The animations should still look natural and smooth
</verification>

<success_criteria>
- Wheel spins for approximately 10 seconds (±1 second acceptable)
- Ball spins for approximately 6 seconds (±0.5 second acceptable)
- Both animations remain smooth and natural-looking
</success_criteria>
