<objective>
Fix several UI issues in the statistics panel on the right side of the roulette wheel game.
</objective>

<context>
This is a Go/Ebitengine roulette wheel game. The stats panel displays game statistics including last number, history, hot/cold numbers, and percentage bars.

File to modify: `./stats/stats.go`
</context>

<requirements>
1. **Last Number section spacing**:
   - The large circle with the last number overlaps/crowds the "Last Number:" label
   - Add more vertical space for this section so subsequent sections (History, Hot Numbers, etc.) are pushed down appropriately
   - Center the number text within its circle (currently appears off-center)

2. **History circles - add numbers**:
   - The History section shows colored circles but no numbers inside
   - Draw the actual number text inside each history circle
   - Use white or contrasting text color so it's readable on red/black/green backgrounds

3. **Hot Numbers circles - add numbers**:
   - Same issue - circles are drawn but numbers are missing
   - Add the number text inside each circle

4. **Cold Numbers circles - add numbers**:
   - Same issue - add number text inside each circle

5. **Even/Odd and Low/High bar colors**:
   - Currently using red/pink colors which don't look good
   - Change to blue color scheme:
     - Dark blue for one side (e.g., Even, Low)
     - Light blue for the other side (e.g., Odd, High)
   - Suggested colors: Dark blue `rgb(30, 60, 120)` and Light blue `rgb(100, 150, 220)` (adjust as needed for visibility)
</requirements>

<implementation>
Look for the drawing functions in stats/stats.go:
- Find where the "Last Number" circle is positioned and adjust Y spacing
- Find where history/hot/cold circles are drawn and add text rendering inside them
- Find the Even/Odd and Low/High bar drawing code and change the colors from red/pink to dark/light blue

For centering text in circles, you'll need to calculate text width and offset appropriately.
</implementation>

<verification>
Build and run with `go build && ./roulette-wheel`:
1. Spin a few times to populate history
2. Verify Last Number circle has proper spacing and centered text
3. Verify History shows numbers inside the colored circles
4. Verify Hot Numbers shows numbers inside circles
5. Verify Cold Numbers shows numbers inside circles
6. Verify Even/Odd bar uses dark blue / light blue
7. Verify Low/High bar uses dark blue / light blue
</verification>

<success_criteria>
- Last Number section has adequate vertical space, number is centered in circle
- All circles (history, hot, cold) display their numbers inside
- Even/Odd and Low/High bars use blue color scheme instead of red/pink
- Text is readable against all background colors
</success_criteria>
