<objective>
Redesign the history display to be a vertical panel on the left side of the screen, showing more results based on research findings.

This improves the UI by giving history more prominence and allowing players to see more past results at a glance.
</objective>

<context>
Read CLAUDE.md for architecture overview.

Key file to modify: `stats/stats.go`
- Currently displays history horizontally in the right-side stats panel
- History limited to 15 items displayed in a row

Research findings: Review `./research/casino-history-displays.md` for casino conventions before implementing.
</context>

<requirements>
1. **New left-side history panel**:
   - Vertical column of past results on the left edge of screen
   - Similar styling to current right-side panel (dark background, gold border)
   - Width appropriate for number chips (~80-100px)

2. **History display**:
   - Show 15+ results (adjust based on research findings)
   - Newest at top, oldest at bottom
   - Each result as a colored chip with number (like current design)
   - Smooth vertical spacing

3. **Layout adjustments**:
   - Wheel should remain centered between the two panels
   - Remove history from the right stats panel (keep other stats there)
   - Ensure responsive to window resizing

4. **Keep existing functionality**:
   - Hot/cold numbers stay in right panel
   - Percentage bars stay in right panel
   - Last number display can stay in right panel or be removed (history shows it)
</requirements>

<implementation>
Modify `stats/stats.go`:
1. Add left panel position/dimensions
2. Create new `DrawHistoryPanel()` method for left side
3. Update `Draw()` to render both panels
4. Update `UpdatePosition()` to handle both panels on resize
5. Remove horizontal history from right panel

Modify `main.go`:
- Update wheel center calculation to account for left panel
- Pass left panel dimensions to stats
</implementation>

<verification>
Build and run:
```bash
go build -o roulette-wheel . && ./roulette-wheel
```

Verify:
- Left panel displays history vertically
- Shows 15+ past results
- Wheel is centered between both panels
- Right panel still shows stats (no history)
- Window resizing works correctly
</verification>

<success_criteria>
- History displayed vertically on left side
- 15+ results visible
- Wheel properly centered
- Clean, casino-like appearance
- No regression in existing stats functionality
</success_criteria>
