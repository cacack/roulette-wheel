# American Roulette Wheel

A fullscreen American roulette wheel application built with Go and Ebitengine for casino night events.

![Screenshot](screenshot.png)

## Features

- Authentic 38-slot American wheel with correct number placement
- Physics-based ball animation with bouncing and settling
- Statistics panel (history, hot/cold numbers, percentages)
- Programmatically generated sound effects
- Fullscreen support

## Build

Requires Go 1.27+. Linux needs [Ebitengine dependencies](https://ebitengine.org/en/documents/install.html).

```bash
go mod tidy
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
| D | Toggle debug overlay (while on, each spin writes a `debug_spin_*.log`) |
| Escape | Exit fullscreen / Exit application |

## Development

```bash
make build      # Build to dist/roulette-wheel
make run        # Run with go run .
make test       # Run all tests, including the statistical fairness simulation
make test-short # Skip the fairness simulation
make clean      # Remove dist/
```

`go run ./cmd/biascheck` simulates 50,000 spins and reports per-number bias. It exits non-zero if spins fail to settle or bias is detected at p < 0.001.

See [CONTRIBUTING.md](CONTRIBUTING.md) for the contribution workflow and [SECURITY.md](SECURITY.md) to report a vulnerability.

## License

See LICENSE file for details.
