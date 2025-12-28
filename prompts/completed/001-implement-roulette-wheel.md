<objective>
Build a fullscreen American roulette wheel application using Go and Ebitengine for use at casino night events. The wheel should provide an authentic visual and audio experience with realistic physics-based animation.

This will be displayed on a large monitor/TV at a casino-themed party, so visual polish and smooth animations are essential for the experience.
</objective>

<context>
- Language: Go with Ebitengine (2D game engine)
- Target: Cross-platform desktop (Windows, macOS, Linux)
- Use case: Casino night entertainment, fullscreen display on monitor
- American roulette: 38 slots (0, 00, 1-36)

Read @CLAUDE.md for project conventions before starting.
</context>

<requirements>
<wheel_design>
- Top-down view of American roulette wheel
- Authentic color scheme: green for 0/00, alternating red/black for 1-36
- Correct number placement following standard American wheel layout:
  0, 28, 9, 26, 30, 11, 7, 20, 32, 17, 5, 22, 34, 15, 3, 24, 36, 13, 1, 00, 27, 10, 25, 29, 12, 8, 19, 31, 18, 6, 21, 33, 16, 4, 23, 35, 14, 2
- Visible number labels on each slot
- Ball track around the outer edge
- Center hub with decorative elements
</wheel_design>

<animation>
- Click anywhere or press Space/Enter to spin
- Wheel rotates in one direction, ball releases in opposite direction
- Ball should initially orbit the outer track at high speed
- Ball gradually slows and drops toward center
- Ball bounces realistically between deflectors/slots before settling
- Final resting position determines the winning number
- Use easing functions for natural deceleration
- Spin duration: 5-10 seconds with variable randomness
</animation>

<statistics_panel>
Position: Right side of wheel (side panel layout)

Display these statistics updated after each spin:
1. **Last Number**: Large, prominent display of most recent result with color indicator
2. **History**: Last 10-15 numbers shown as colored chips/circles
3. **Hot Numbers**: Top 3-5 most frequently hit numbers
4. **Cold Numbers**: 3-5 least frequently hit numbers
5. **Red/Black %**: Running percentage with visual bar
6. **Even/Odd %**: Running percentage with visual bar
7. **High/Low %**: Running percentage (1-18 vs 19-36)
</statistics_panel>

<sound_effects>
- Spinning wheel: clicking/ticking sound that matches rotation speed
- Ball bouncing: realistic ball-on-wood/metal sounds
- Ball settling: final drop sound
- Win announcement: subtle chime when ball stops
- Sounds should be toggleable with 'M' key for mute
</sound_effects>

<controls>
- Click anywhere or Space/Enter: Spin the wheel
- F or F11: Toggle fullscreen
- M: Toggle sound mute
- R: Reset statistics
- Escape: Exit application
</controls>
</requirements>

<implementation>
<structure>
Organize code into logical packages:
- `main.go` - Entry point, game loop
- `wheel/` - Wheel rendering and physics
- `ball/` - Ball physics and animation
- `stats/` - Statistics tracking and display
- `audio/` - Sound effect management
- `assets/` - Embedded sound files (use Go embed)
</structure>

<technical_notes>
- Use Ebitengine's vector graphics or pre-rendered assets for the wheel
- Implement proper game loop with Update/Draw separation
- Use deterministic randomness for the final position (crypto/rand for fairness)
- Sound files can be simple WAV or OGG format
- Consider generating simple sounds programmatically if asset creation is complex
- Target 60 FPS for smooth animation
- Handle window resize gracefully while maintaining wheel proportions
</technical_notes>

<avoid>
- Don't overcomplicate the physics - visual believability matters more than perfect simulation
- Don't block the main thread with sound loading - load assets asynchronously
- Avoid magic numbers - define constants for wheel layout, colors, timing
</avoid>
</implementation>

<output>
Create the complete application with this structure:
- `./main.go` - Application entry point
- `./go.mod` - Go module definition
- `./wheel/wheel.go` - Wheel rendering
- `./ball/ball.go` - Ball physics
- `./stats/stats.go` - Statistics tracking
- `./audio/audio.go` - Sound management
- `./assets/` - Any required asset files

Include a README.md with:
- Build instructions for each platform
- Controls reference
- Screenshot placeholder
</output>

<verification>
Before declaring complete:
1. Run `go build` - must compile without errors
2. Run the application - wheel should display centered with stats panel on right
3. Click to spin - wheel and ball should animate smoothly
4. Verify sound plays during spin (if not muted)
5. Confirm statistics update after each spin
6. Test fullscreen toggle with F key
7. Verify the wheel uses correct American roulette number sequence
</verification>

<success_criteria>
- Application compiles and runs on the current platform
- Wheel displays with correct 38-slot American layout
- Clicking triggers a realistic spin animation with counter-rotating ball
- Ball settles into a random slot after 5-10 seconds
- Statistics panel updates correctly after each spin
- Sound effects enhance the experience
- Fullscreen mode works properly
- Smooth 60 FPS animation throughout
</success_criteria>
