# American Roulette Wheel

A fullscreen American roulette wheel application built with Go and Ebitengine for casino night events. Features realistic physics-based animation, authentic wheel design, and comprehensive statistics tracking.

## Features

- Authentic American roulette wheel with 38 slots (0, 00, 1-36)
- Correct number placement following standard American wheel layout
- Realistic spinning animation with physics-based ball movement
- Ball bounces between deflectors before settling
- Statistics panel tracking:
  - Last number with color indicator
  - History of last 15 numbers
  - Hot and cold numbers
  - Red/Black, Even/Odd, High/Low percentages
- Programmatically generated sound effects
- Fullscreen support for large displays
- Smooth 60 FPS animation

## Build Instructions

### Prerequisites

- Go 1.21 or later
- For Linux: Additional dependencies for Ebitengine (see [Ebitengine installation guide](https://ebitengine.org/en/documents/install.html))

### Building

```bash
# Clone the repository
git clone <repository-url>
cd roulette-wheel

# Download dependencies
go mod tidy

# Build the application
go build -o roulette-wheel .
```

### Platform-Specific Notes

**macOS:**
```bash
go build -o roulette-wheel .
./roulette-wheel
```

**Windows:**
```bash
go build -o roulette-wheel.exe .
roulette-wheel.exe
```

**Linux:**
```bash
# Install dependencies first (Ubuntu/Debian)
sudo apt-get install libc6-dev libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config

go build -o roulette-wheel .
./roulette-wheel
```

## Controls

| Key | Action |
|-----|--------|
| Click / Space / Enter | Spin the wheel |
| F / F11 | Toggle fullscreen |
| M | Toggle mute |
| R | Reset statistics |
| Escape | Exit fullscreen / Exit application |

## Project Structure

```
roulette-wheel/
├── main.go           # Application entry point and game loop
├── wheel/
│   └── wheel.go      # Wheel rendering and layout
├── ball/
│   └── ball.go       # Ball physics and animation
├── stats/
│   └── stats.go      # Statistics tracking and display
├── audio/
│   └── audio.go      # Sound effect management
├── go.mod            # Go module definition
└── README.md         # This file
```

## American Roulette Wheel Layout

The wheel uses the standard American roulette number sequence (clockwise from 0):

```
0, 28, 9, 26, 30, 11, 7, 20, 32, 17, 5, 22, 34, 15, 3, 24, 36, 13, 1, 00, 27, 10, 25, 29, 12, 8, 19, 31, 18, 6, 21, 33, 16, 4, 23, 35, 14, 2
```

- Green: 0, 00
- Red: 1, 3, 5, 7, 9, 12, 14, 16, 18, 19, 21, 23, 25, 27, 30, 32, 34, 36
- Black: 2, 4, 6, 8, 10, 11, 13, 15, 17, 20, 22, 24, 26, 28, 29, 31, 33, 35

## License

See LICENSE file for details.
