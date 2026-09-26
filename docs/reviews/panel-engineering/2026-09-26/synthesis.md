# Engineering Panel Synthesis — 2026-09-26

## Per-persona verdicts
| Persona | Verdict | Findings (C/H/M/L) |
|---------|---------|--------------------|
| Architect | needs-attention | 0/2/2/1 |
| Security Posture | needs-attention | 0/2/1/3 |
| Operations/SRE | needs-attention | 0/0/1/2 |
| Developer Experience | needs-attention | 0/1/3/1 |
| Maintainability | needs-attention | 0/1/2/2 |
| **Total** | | **0/6/9/9** |

## Cross-cutting themes

1. **No automated verification of outcome-determining code** — *architect (H), maintainability (H), ops-sre (M), dx (M)*.
   Zero `*_test.go` across ~3.8k Go LOC; no `make test`; `build.yml` gates tagged releases on vet/govulncheck/build only. `cmd/biascheck` (chi-square fairness simulation) exists but has no assertions and never runs in CI. The strongest signal in the review: four of five personas raised it independently.

2. **Undocumented tooling and behavior** — *dx (M×2), ops-sre (L), maintainability (M)*.
   Makefile not mentioned in README; `cmd/biascheck` absent from README and CLAUDE.md; `D` debug overlay and `debug_spin_*.log` writer absent from README controls; CLAUDE.md says results use crypto/rand but `main.go:755` also feeds a fixed-seed LCG (`prngState = 42`) into initial wheel speed.

3. **`prompts/` partially tracked** — *security (M), maintainability (M), dx (L)*.
   `.gitignore` excludes `prompts/`, yet 11 of 23 files under `prompts/completed/` remain tracked from before the ignore rule. No secrets found, but the tracked set is inconsistent and unexplained.

4. **Open-source contributor and disclosure surface missing** — *security (H), dx (H)*.
   No SECURITY.md and no CONTRIBUTING.md / issue or PR templates, although the repo is public and carries `good first issue` / `help wanted` labels.

## Prioritized findings

1. **[H] No tests; fairness check not in CI** — architect, maintainability, ops-sre, dx. `ball/ball.go`, `wheel/wheel.go`, `.github/workflows/build.yml`, `cmd/biascheck/main.go`.
2. **[H] Number classification duplicated three times** — architect. `wheel` exports `RedNumbers`/`IsRed`/`IsEven`/`IsLow`, but `main.go:425` and `stats/stats.go:46,99,464` maintain their own copies.
3. **[H] Dependabot does not cover `gomod`** — security (ops-sre noted). `.github/dependabot.yml` tracks only `github-actions`.
4. **[H] No SECURITY.md** — security.
5. **[H] No contribution path** — dx.
6. **[M] `prompts/` partially tracked** — security, maintainability, dx.
7. **[M] Second, predictable randomness source undocumented** — maintainability. `main.go:755`.
8. **[M] Dev tooling (Makefile, biascheck, debug overlay) undiscoverable** — dx, ops-sre.
9. **[M] `wheel/wheel.go` mixes domain state with ~40 rendering helpers** — architect.
10. **[M] `Game` in `main.go` is the sole coordinator, with direct file I/O** — architect.

Low items (not drafted): no SAST/secret scanning in CI; no documented release rollback; debug-log destination not configurable; `main.go:307` untagged "could implement" stub; a local absolute path embedded in tracked `research/audio-options.md` (privacy: it should be replaced with a repo-relative path).

## Overall assessment

The repo is small, well structured and well kept. It has an acyclic package graph, accurate CLAUDE.md architecture notes, SHA-pinned Actions, crypto/rand for outcomes, no network or exec surface, and no secrets. The main risk is that the code which decides where the ball lands has no automated checks, while a fairness tool already exists and is simply not wired into CI. Fix that first: add unit tests for `ball` and `wheel`, turn `biascheck` into an assertion that CI runs, and add `go test` and `make test`. Next, consolidate the duplicated red/black classification into `wheel`, which is cheap and removes silent drift. The remaining items are hygiene for a public repo: gomod Dependabot, SECURITY.md, CONTRIBUTING.md, cleaning up `prompts/`, and documenting the tooling.
