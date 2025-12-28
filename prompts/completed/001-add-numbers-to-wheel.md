<objective>
Add visible number text to the American roulette wheel slots.

Currently, the wheel renders colored circles as placeholders instead of actual numbers (see `drawNumberIndicator` function in `wheel/wheel.go:313-339`). Replace this with proper text rendering showing the correct numbers (0, 00, 1-36) in their respective slots.
</objective>

<context>
This is a Go application using Ebiten (github.com/hajimehoshi/ebiten/v2) for graphics.

Key files:
- `wheel/wheel.go` - Contains wheel rendering logic, including the problematic `drawNumberIndicator` function
- The `NumberSequence` array (line 13-18) defines the correct American roulette layout

The wheel already has:
- Correct slot colors (red, black, green)
- Correct number sequence for American roulette
- Slot dividers and chrome accents

Review `@CLAUDE.md` for project conventions.
</context>

<requirements>
1. Render actual number text (0, 00, 1-36) in white on each slot
2. Numbers should be readable and properly positioned in the center of each slot
3. Text should rotate with the wheel (numbers stay oriented relative to their slot)
4. Use Ebiten's text rendering capabilities (likely `text/v2` package or similar)
5. Font size should be appropriate for the slot size - visible but not overflowing

American roulette number sequence (clockwise from 0):
0, 28, 9, 26, 30, 11, 7, 20, 32, 17, 5, 22, 34, 15, 3, 24, 36, 13, 1, 00, 27, 10, 25, 29, 12, 8, 19, 31, 18, 6, 21, 33, 16, 4, 23, 35, 14, 2
</requirements>

<implementation>
1. Add necessary imports for Ebiten text rendering
2. Initialize a font face (can use Go's built-in fonts or Ebiten's default)
3. Replace the `drawNumberIndicator` function to render actual text
4. Position text at the center of each slot at the correct angle
5. Use white color for text to contrast against red/black/green backgrounds

Note: The existing `drawNumberIndicator` draws colored circles as placeholders. Replace this approach entirely with text rendering.
</implementation>

<output>
Modify: `./wheel/wheel.go`
</output>

<verification>
1. Run `go build` to ensure compilation succeeds
2. Run the application and visually verify:
   - All 38 numbers are visible on the wheel
   - Numbers are in the correct American roulette sequence
   - Numbers rotate with the wheel
   - Text is legible against slot backgrounds
</verification>

<success_criteria>
- Numbers 0, 00, and 1-36 are all visible on the wheel
- Numbers match the standard American roulette layout
- Text is white and readable against the colored slot backgrounds
- Application compiles and runs without errors
</success_criteria>
