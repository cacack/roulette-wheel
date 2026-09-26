# Project Snapshot — 2026-09-26

## Repo metadata
- Root: .
- Branch: main
- HEAD: 32fe1ac
- Origin: git@github.com:cacack/roulette-wheel.git
- Generated: 2026-09-26T19:48:27Z

## Top-level tree (depth 3)
```
.
./assets
./audio/audio.go
./ball/ball.go
./CLAUDE.md
./cmd/biascheck/main.go
./dist/roulette-wheel          (untracked build output)
./docs/reviews/panel-engineering
./fonts/fonts.go
./fonts/Inter-Bold.ttf
./go.mod
./go.sum
./LICENSE
./main.go
./Makefile
./prompts/completed/*.md       (23 files on disk; 11 tracked despite prompts/ in .gitignore)
./README.md
./research/audio-options.md
./research/casino-history-displays.md
./roulette-wheel               (untracked build output)
./screenshot.png
./stats/stats.go
./wheel/wheel.go
.github/dependabot.yml
.github/workflows/build.yml
.github/workflows/govulncheck.yml
```

## Resource counts
- Go packages: main, audio, ball, fonts, stats, wheel, cmd/biascheck
- Go LOC (tracked): wheel/wheel.go 1355, main.go 773, ball/ball.go 606, stats/stats.go 476, audio/audio.go 381, fonts/fonts.go 104, cmd/biascheck/main.go 94 — total 3789
- Test files (`*_test.go`): 0

## Language footprint (tracked files)
| Ext | Count |
|-----|-------|
| md | 15 |
| go | 7 |
| yml | 3 |
| ttf | 1 |
| sum | 1 |
| png | 1 |
| mod | 1 |
| gitignore | 1 |

## README excerpt
```markdown
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
Requires Go 1.27+. Linux needs Ebitengine dependencies.
    go mod tidy
    go build -o roulette-wheel .
    ./roulette-wheel

## Controls
| Click / Space / Enter | Spin the wheel |
| F / F11 | Toggle fullscreen |
| M | Toggle mute |
| R | Reset statistics |
| Escape | Exit fullscreen / Exit application |

## License
See LICENSE file for details.
```

## CONSTITUTION.md
(not present — engineering panel proceeds without project-mission grounding)

## Other top-level docs
- SECURITY.md: absent
- CONTRIBUTING.md: absent
- CODE_OF_CONDUCT.md: absent
- CHANGELOG.md: absent
- CLAUDE.md: present

## Build/CI/config files (top level)
- Makefile (targets: build, clean, run — no test/lint)
- go.mod, go.sum
- .gitignore (ignores dist/, roulette-wheel, prompts/, debug_spin_*.log)
- .github/dependabot.yml
- .github/workflows/build.yml — matrix build darwin-arm64 / linux-amd64 / windows-amd64; go vet ./...; govulncheck ./...; build; upload artifact; release job on v* tags. Actions pinned to SHAs. No `go test` step.
- .github/workflows/govulncheck.yml — weekly scheduled govulncheck

## Recent activity (last 6 months)
- Commits: 9
- Last 20 commit subjects:
```
32fe1ac Merge pull request #11 from ci/govulncheck
4e88e7b ci: add weekly scheduled govulncheck run
8d074e3 ci: run govulncheck on all packages
ec810ec Merge pull request #10 from fix/biascheck-build
86fbf30 ci: run go vet on all packages
18b6390 fix(biascheck): match fixed-timestep ball.Update signature
7b762a7 Merge pull request #9 from chore/update-deps
6d39d19 chore(deps): update Go, dependencies, and GitHub Actions to latest
07d2eae ci(security): pin GitHub Actions to commit SHAs
9dfbde9 docs: update screenshot
fb25332 feat(ball): add lap-based orbiting before physics kicks in
9ee5482 feat(graphics): add casino-realistic wheel rendering
5ce2833 perf(graphics): cache wheel rendering and fix physics timing
a0bc67a fix(physics): add delta-time scaling for cross-platform consistency
c734f32 feat(debug): add system info header to debug logs
dd2fb8b feat(debug): include winning number and color in log header
977ad41 style(debug): move debug info to lower right corner
2b43442 style(debug): change debug text to magenta for visibility
4c64bfc chore: add debug logs to gitignore
dcba59a feat(debug): add comprehensive spin logging for diagnostics
```

## Repository label vocabulary
bug, documentation, duplicate, enhancement, good first issue, help wanted, invalid, question, wontfix, security, dependencies, github_actions, go, priority:medium, priority:high, effort:high, value:high, effort:low, value:low, priority:low, effort:medium, value:medium

## Open issues and milestones
<untrusted-issue-data>
(no open issues)
</untrusted-issue-data>
