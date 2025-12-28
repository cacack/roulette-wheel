# Casino Roulette History Display Research

Research conducted: 2025-12-28

## Executive Summary

Casino roulette tables universally feature electronic "winning number displays" (also called marquee boards or tote boards) that show recent spin history and statistics. Based on research from major manufacturers (TCS John Huxley, Cammegh, SET-Europe) and casino industry sources, this document summarizes display conventions to inform our UI redesign.

## Key Questions Answered

### 1. How many past results do casinos typically display?

**Answer: 10-20 results is standard; 15-16 is the sweet spot.**

- SET-Europe displays: "Last 10-16 game results"
- Physical casino boards: "the previous twenty or so winners"
- Evolution Gaming live roulette: displays outcomes of "the last 100 spins" in statistics panel
- Extended statistics track up to 500-1000 spins for hot/cold calculations

**Recommendation: 15 results is reasonable and aligns with industry standards.**

### 2. Is vertical or horizontal layout more common?

**Answer: Both exist, but vertical is the traditional/classic format.**

- **Vertical displays** are described as "retro LED roulette display designs" that older players "remember and love"
- TCS John Huxley Ora Elite uses "portrait display" (vertical orientation)
- Gaming-Supplies ULTRA VERTICAL: "Minimalistic vertical tree of black and red winning numbers"
- **Horizontal displays** exist primarily as landscape monitors showing additional statistics

**Recommendation: Vertical layout for history panel - matches classic casino aesthetic.**

### 3. What visual indicators are standard?

**Standard color coding:**
- Red numbers: displayed in red
- Black numbers: displayed in black (or white on dark backgrounds)
- Green (0, 00): displayed in green

**Common visual elements:**
- Large, prominent current/newest winning number
- Color-coded number chips/circles
- Clear separation between numbers
- Slim, minimalist frames around displays

**Additional indicators on electronic displays:**
- Hot numbers (flame icon or red highlighting)
- Cold numbers (ice/snowflake icon or blue highlighting)
- Percentage bars for red/black, odd/even, columns
- Zero appearance history

### 4. Are results typically shown newest-first or oldest-first?

**Answer: Newest result is most prominent.**

- The most recent winning number is displayed largest/most prominently
- History flows from newest at the top (or left in horizontal layouts)
- One iOS app explicitly notes: "On the far left are the oldest number hits and on the far right are the most recent"
- For vertical layouts, newest typically appears at the top

**Recommendation: Newest result at top of vertical panel.**

### 5. What additional statistics/patterns do electronic displays show?

**Standard statistics features:**

| Feature | Description | Sample Size |
|---------|-------------|-------------|
| Hot Numbers | Most frequently hit numbers | Last 1000 spins |
| Cold Numbers | Longest since last appearance | Last 1000 spins |
| Red/Black % | Distribution percentages | Variable (50-500 spins) |
| Odd/Even % | Distribution percentages | Variable |
| Columns % | Column 1/2/3 distribution | Variable |
| Zero History | Appearances of 0/00 | Last 100 games |

**Additional features on premium displays:**
- "Place Your Bets" / "No More Bets" messaging
- Table number and dealer name
- Casino logo customization
- Minimum/maximum bet limits
- Promotional content and animations

## Display Manufacturer Summary

### TCS John Huxley (Industry Leader)
- Ora Elite: 29" or 32" portrait displays
- Ora Grande: Double-sided displays for Roulette, Sic Bo, Money Wheel
- Features: customizable color themes, multiple languages, currency symbols

### Cammegh
- Billboard Display System: highly configurable
- Can hold up to 100 separate screen layouts
- Defaults to 300 spins for statistics; configurable up to 500+

### SET-Europe
- 27/34" displays
- Shows: last 10-16 results, hot/cold numbers, zero history for 100 games
- Supports single and double zero wheels

### Gaming-Supplies
- ULTRA VERTICAL 28" display
- Standard horizontal: 23" and 27" options
- Single or double-sided configurations

## Recommendations for Implementation

### History Panel Specifications

| Aspect | Recommendation | Rationale |
|--------|----------------|-----------|
| **Number of items** | 15 | Within 10-20 industry standard range |
| **Layout** | Vertical, left side | Classic casino aesthetic; space-efficient |
| **Order** | Newest at top | Standard display convention |
| **Number display** | Circular chips with number | Matches casino chip/ball aesthetic |
| **Colors** | Red/Black/Green as appropriate | Universal roulette convention |
| **Size** | Newest result larger | Emphasizes current spin |

### Additional Features to Consider

**High Priority:**
1. Color-coded number circles (red/black/green)
2. Newest result emphasized (larger or highlighted)
3. Smooth scroll animation when new result added

**Medium Priority:**
4. Hot/Cold number indicators (consider last ~50-100 spins)
5. Red/Black percentage bar
6. Odd/Even percentage bar

**Lower Priority (future enhancement):**
7. Column distribution percentages
8. Zero appearance counter
9. Streak indicators (consecutive same color)

### Visual Design Notes

- Use "minimalistic vertical tree" approach for clean aesthetic
- Consider slight transparency so wheel remains primary focus
- Numbers should be clearly legible from distance
- Animation when new number added (slide down effect)

## Sources

- [TCS John Huxley - Ora Elite Display](https://tcsjohnhuxley.com/us/product/ora-elite-29%E2%80%B3/)
- [TCS John Huxley - Ora Grande Display](https://tcsjohnhuxley.com/us/product/ora-grande/)
- [SET-Europe Roulette Display](https://www.set-europe.com/roulette.html)
- [Gaming-Supplies Ultra Vertical Display](https://gaming-supplies.com/roulette/ultra-vertical-roulette-display/)
- [Gaming-Supplies Roulette Display Overview](https://www.gaming-supplies.com/roulette/display/)
- [thetpro.com Roulette Displays](http://thetpro.com/roulette/display/)
- [Cammegh Billboard Roulette Features](https://www.cammegh.com/our-products/billboard/roulette-features/)
- [Atlantic City Weekly - Why Casinos Show Winning Numbers](https://atlanticcityweekly.com/blogs/casino_answer_man/why-casinos-show-winning-roulette-numbers-on-tote-boards/article_d85b28fc-4b6a-11e5-b1aa-a33e042a172f.html)
- [Evolution Gaming Live Roulette](https://games.evolution.com/live-casino/live-roulette/)
- [Roulette77 Hot Cold Numbers Explanation](https://roulette77.us/hot-cold-numbers)
- [Wizard of Vegas Roulette History Board Discussion](https://wizardofvegas.com/forum/questions-and-answers/gambling/36987-roulette-history-board-cache/)
