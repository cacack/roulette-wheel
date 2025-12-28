<objective>
Fix the roulette wheel physics to match real-world behavior:
1. The wheel spins longer than the ball (wheel ~10 sec, ball ~6 sec)
2. Ball settles while wheel is still slowly spinning
3. Wheel continues for 3-5 seconds AFTER ball settles, then stops
4. Result is declared only AFTER the wheel has completely stopped
5. The winning number must be calculated from the ball's actual position on the stopped wheel, NOT pre-determined randomly
</objective>

<context>
This is a roulette wheel simulation built with Go and Ebitengine (60fps).

Current broken behavior:
- `TargetSlot` is picked RANDOMLY at spin start (`ball/ball.go:89`)
- Ball animates toward this pre-determined slot
- Result is declared when ball settles, even if wheel is still moving
- The declared number can visually "jump" because the wheel continues rotating after declaration

Real roulette physics:
- The wheel is heavy and spins for a long time with gradual slowdown
- The ball orbits, loses speed, drops into the wheel slots
- The ball bounces between slots as the wheel continues to spin
- Eventually the ball settles in a slot while the wheel is still moving slowly
- The wheel gradually stops, and only THEN is the result final
- The winning number depends on WHERE the ball is when the wheel stops
</context>

<requirements>
1. **Remove pre-determined TargetSlot**: Don't pick the winning slot at spin start. Let physics determine where the ball lands.

2. **Physics-based ball settling**: The ball should:
   - Orbit and slow down based on friction
   - Drop into the slot area when speed is low enough
   - Bounce/settle based on wheel's current rotation and speed
   - Lock into a slot naturally

3. **Correct timing sequence**:
   - Ball spins for ~6 seconds, then settles into a slot
   - Wheel continues spinning for 3-5 more seconds after ball settles
   - Wheel gradually stops
   - ONLY when wheel stops: calculate and declare the winning number

4. **Result calculation**: When the wheel has stopped, determine which slot the ball is in by:
   - Getting the ball's current angle
   - Getting the wheel's final rotation
   - Calculating: `slotIndex = floor((ballAngle - wheelRotation) / SlotAngle) mod NumSlots`
   - Looking up `NumberSequence[slotIndex]`

5. **Visual consistency**: The declared number must EXACTLY match the slot the ball is visually sitting in.
</requirements>

<implementation>
Files to modify:

**`./ball/ball.go`**:
- Remove `TargetSlot` random selection from `StartSpin()`
- Update settling logic to find the nearest slot based on current position
- Ball should lock to the nearest slot when settling, not a pre-determined one

**`./main.go`**:
- Adjust timing: wheel ~10 sec, ball ~6 sec
- Move result declaration from `OnSettle` callback to AFTER wheel stops
- Add wheel stop detection and result calculation there
- Calculate winning slot from ball's actual position when wheel is stationary

**`./wheel/wheel.go`** (if needed):
- Ensure wheel friction allows it to spin 3-5 seconds after ball settles

Key constants (at 60fps):
- Ball duration: ~360 ticks (6 sec)
- Wheel duration: ~540-600 ticks (9-10 sec, so 3-4 sec after ball settles)
</implementation>

<verification>
Build and run with `go build && ./roulette-wheel`:
1. Spin the wheel multiple times
2. Observe: ball should settle while wheel is still moving
3. Wheel should continue spinning for a few seconds after ball settles
4. Result should be declared only when wheel stops
5. The declared number must match the slot the ball is visually in - NO jumping
6. Run 10+ spins to confirm consistent behavior
</verification>

<success_criteria>
- Ball settles while wheel is still moving (wheel outlasts ball by 3-5 seconds)
- Result is declared only after wheel completely stops
- Declared number exactly matches the ball's visual position every time
- No visual "jumping" of the result
- Physics feel natural and realistic
</success_criteria>
