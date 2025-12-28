<objective>
Fix cross-platform speed inconsistency and increase ball/wheel speeds for realistic casino feel.

The wheel and ball run significantly slower on Windows than on Mac. Additionally, regardless of platform, the ball should spin much faster (like it's whizzing by as in a real casino) and the wheel should spin moderately faster.
</objective>

<context>
Read CLAUDE.md for architecture overview.

Key files:
- `main.go` - Contains `WheelSpinSpeed` and `WheelFriction` constants
- `ball/ball.go` - Contains `InitialAngularSpeed` and friction constants

Ebitengine runs at 60 TPS (ticks per second) set via `ebiten.SetTPS(TargetFPS)`. Physics should be frame-rate independent but may not be currently.
</context>

<requirements>
1. **Ball speed increase**: Approximately 300% faster initial angular speed
   - Current `InitialAngularSpeed = 0.15` should become ~0.45
   - Ball should appear to whiz around the track like in a real casino

2. **Wheel speed increase**: Approximately 150-200% faster initial spin
   - Current `WheelSpinSpeed = 0.05` should become ~0.075-0.10
   - Wheel should have more momentum when spun

3. **Cross-platform consistency**: Ensure speeds are consistent between Windows and Mac
   - Verify physics calculations use delta time or are frame-rate independent
   - If using fixed TPS (60), speeds should already be consistent - investigate why they differ

4. **Maintain spin duration**: Total spin time should remain roughly the same (~5.5-6.5 seconds)
   - Adjust friction coefficients proportionally if needed to maintain duration
   - Higher speeds with higher friction = same duration but more exciting visual
</requirements>

<implementation>
Approach:
1. Increase initial speeds as specified
2. Proportionally increase friction to maintain overall spin duration
3. Test that the feel is more exciting without changing timing significantly

Key constants to adjust:
- `ball/ball.go`: `InitialAngularSpeed`, `FrictionCoefficient`
- `main.go`: `WheelSpinSpeed`, `WheelFriction`, `WheelBrakeFriction`
</implementation>

<verification>
Build and run the application:
```bash
go build -o roulette-wheel . && ./roulette-wheel
```

Verify:
- Ball visibly whizzes around the outer track (feels fast/exciting)
- Wheel has noticeable spin momentum
- Total spin duration is still approximately 5.5-6.5 seconds
- Animation feels smooth, not jerky
</verification>

<success_criteria>
- Ball initial speed increased ~300% (0.15 → ~0.45)
- Wheel initial speed increased ~150-200% (0.05 → ~0.075-0.10)
- Spin duration remains approximately the same
- Animation appears smooth and casino-like
</success_criteria>
