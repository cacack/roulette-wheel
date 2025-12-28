<objective>
Add a dramatic winning number reveal animation when the ball settles.

When the ball lands on a number, display a large colored circle filling most of the screen showing the winning number, then animate it shrinking down to the "Last Number" position in the stats panel. This creates excitement and clearly announces the result.
</objective>

<context>
Read CLAUDE.md for architecture overview.

Key files:
- `main.go` - Game loop, handles result declaration in `declareResult()`
- `stats/stats.go` - Stats panel rendering, includes `drawLastNumber()` method

The result is declared when `declareResult()` is called after wheel stops and ball has settled.
</context>

<requirements>
1. **Full-screen reveal animation**:
   - When ball settles and result is declared, show large colored circle
   - Circle should be centered on screen, sized to ~80% of screen height
   - Display winning number in large text, centered in circle
   - Color matches the number (red/black/green)
   - Hold for ~2 seconds

2. **Shrink animation**:
   - After hold, animate circle shrinking
   - Move toward the "Last Number" position in stats panel
   - Smooth easing (ease-out or similar)
   - Duration ~1 second

3. **Larger last number display**:
   - Increase the existing "Last Number" chip size in stats panel
   - Current `chipSize := 40.0` - increase to ~60-70
   - Adjust text scaling accordingly

4. **Animation state management**:
   - Add animation phase tracking (revealing, shrinking, done)
   - Track animation progress (0.0 to 1.0)
   - Reset animation state when new spin starts
</requirements>

<implementation>
Approach:
1. Add animation state fields to Game struct (phase, progress, target position)
2. Create animation update logic in `Update()` after result declared
3. Add `drawWinningAnimation()` method to render the animated circle
4. Call it in `Draw()` when animation is active (draw on top of everything)
5. Increase chip size in `drawLastNumber()`

Animation timing:
- Hold phase: ~120 frames (2 seconds at 60fps)
- Shrink phase: ~60 frames (1 second at 60fps)

Use easing function for smooth shrink:
```go
// Ease-out cubic
progress := animationProgress // 0.0 to 1.0
eased := 1 - math.Pow(1-progress, 3)
```
</implementation>

<verification>
Build and run:
```bash
go build -o roulette-wheel . && ./roulette-wheel
```

Verify:
- Spin the wheel and watch ball settle
- Large colored circle appears with winning number
- Circle holds for ~2 seconds
- Circle smoothly shrinks to stats panel position
- Last number display is larger than before
- Animation resets on next spin
</verification>

<success_criteria>
- Full-screen reveal animation displays on win
- Smooth shrink animation to stats panel
- Last number chip size increased
- No visual glitches or timing issues
- Animation doesn't interfere with next spin
</success_criteria>
