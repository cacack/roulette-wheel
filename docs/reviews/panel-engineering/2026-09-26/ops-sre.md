# Operations / SRE Review — 2026-09-26

**Verdict:** needs-attention

**Project scale:** personal / hobby desktop application — a fullscreen Ebitengine roulette wheel run locally for casino-night events. No server, no production deploy target, no network dependencies. "Operations" here means: can someone at an event build/run it, and can a maintainer ship and diagnose it without surprises.

This is not a service, so most classic SRE concerns (health checks, circuit breakers, metrics scraping, on-call rotation) genuinely do not apply, and their absence is not a finding. What does apply at this scale is: does the release pipeline give any confidence before code reaches a downloadable binary, and can someone troubleshoot a bad run at the venue with what's on hand. On those two axes there are real, scale-appropriate gaps worth closing, but nothing here is actively dangerous.

## Findings

**[MEDIUM] Release pipeline has no automated test gate**
- Evidence: `.github/workflows/build.yml` runs `go vet ./...`, `govulncheck ./...`, then `go build`, straight into an artifact upload and (on `v*` tags) a GitHub Release publish. Snapshot confirms zero `*_test.go` files in the repo (3789 tracked Go LOC, 0 tests).
- Why it matters: the only thing gating a tagged release reaching a downloadable binary is "it compiles and passes static analysis." A physics/statistics regression (e.g., in `ball/ball.go`'s phase state machine or `stats/stats.go` percentage math) can ship straight to a release artifact with no automated signal, only manual play-testing after the fact.
- Suggested action: even a handful of table-driven tests around deterministic logic (e.g., `wheel.NumberSequence`/`RedNumbers` lookups, `stats` percentage math) plus a `go test ./...` step in `build.yml` would give a cheap pre-release check without requiring UI/physics test infrastructure.

**[LOW] Debug/diagnostic mode is undocumented**
- Evidence: pressing `D` toggles `g.showDebug` (main.go:296-298), which turns on an on-screen overlay and, on spin end, writes a per-spin timing log via `writeDebugLog` (main.go:538-589) to `debug_spin_<timestamp>.log` in the working directory (gitignored per `.gitignore`). The README's Controls table (Click/Space/Enter, F/F11, M, R, Escape) does not mention `D`.
- Why it matters: this is the app's only built-in diagnostic tool — exactly what someone would reach for if the wheel behaves oddly at a live event (the commit history shows it was built for cross-platform timing bugs: `a0bc67a fix(physics): add delta-time scaling for cross-platform consistency`, `c734f32`/`dd2fb8b feat(debug): add system info/winning number to debug logs`). Without it being documented anywhere, whoever is running the kiosk at an event has no way to know the tool exists or where the resulting log file goes.
- Suggested action: add a line to the README Controls table (`D — Toggle debug overlay / per-spin diagnostic log`) so the one piece of operational tooling this app has is discoverable.

**[LOW] No documented recovery path for a bad tagged release**
- Evidence: `build.yml`'s `release` job runs unconditionally on any `v*` tag push, downloads all matrix artifacts, and publishes a GitHub Release with `generate_release_notes: true` — fully automatic, no manual approval gate. No `docs/`, `RUNBOOK.md`, or README section describes what to do if a bad build is released.
- Why it matters: low likelihood/low blast-radius for a hobby project, but if a broken binary is tagged and published, there's no written guidance on the fix-forward vs. delete-and-retag choice (deleting a GitHub release/tag is a mildly destructive action worth having a stated preference for before someone's improvising at an event weekend).
- Suggested action: a one-paragraph note in the README or a short `RELEASING.md` ("to cut a release, push a `vX.Y.Z` tag; to fix a bad release, delete the tag/release and re-tag" or similar) would remove any ambiguity, if the maintainer thinks it's worth codifying.

## Notes
- Health checks, readiness probes, timeouts/retries, circuit breakers, and log aggregation are all not applicable — this is a locally-run GUI application with no external service dependencies and no server component. Their absence is expected, not a gap.
- `log.Printf`/`log.Fatal` usage (main.go:557, 588, 771) is minimal but proportionate: two diagnostic prints and one fatal-on-startup-error for `ebiten.RunGame`. No structured logging is warranted at this scale.
- Multi-platform build matrix (darwin-arm64, linux-amd64, windows-amd64) in `build.yml` with pinned action SHAs is a solid, appropriately-sized reproducibility story for a binary-distribution project — no lockfile/base-image concerns apply since there's no container image being shipped.
- `dependabot.yml` only tracks the `github-actions` ecosystem, not `gomod` — Go module updates rely on the manual `chore(deps): update ...` commits seen in history (e.g., `6d39d19`). This is adjacent to supply-chain/security scope rather than pure operability, so it's noted here rather than raised as a finding.
- No runbooks/`ops/` directory exists, and none is warranted — there's no incident surface to document beyond what's covered above.

### Summary counts
critical=0 high=0 medium=1 low=2
