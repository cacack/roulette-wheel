# Proposed Issues — 2026-09-26

## 1. Add unit tests and run the biascheck fairness test in CI
**Severity:** high  **Persona(s):** architect, maintainability, ops-sre, dx  **Labels:** enhancement, priority:high

The repo has no `*_test.go` files across roughly 3.8k Go LOC, and `ball/ball.go` and `wheel/wheel.go` decide every outcome with nothing checking them automatically. `cmd/biascheck` already simulates spins to check the per-slot distribution, but it has no pass/fail assertion and never runs in CI. `.github/workflows/build.yml` publishes release artifacts on `v*` tags after only vet, govulncheck and build.

**Suggested approach:**
- Add table-driven tests for the `ball` phase transitions and settling, and for the `wheel` slot/angle mapping.
- Move the biascheck core into a test with a chi-square threshold (skipped under `-short` if it is slow), or give it a non-zero exit on bias.
- Add a `test` target to the Makefile and a `go test ./...` step to `build.yml` before the build.

---

## 2. Consolidate roulette number classification into the wheel package
**Severity:** high  **Persona(s):** architect  **Labels:** enhancement, effort:low

`wheel` exports `RedNumbers`, `GetNumberColor`, `IsRed`, `IsEven` and `IsLow`, but `main.go:425` and `stats/stats.go:46,99,464` each keep their own `redNumbers` map and parity/range logic. The copies agree today by coincidence rather than by construction.

**Suggested approach:** replace the local copies in `main.go` and `stats` with calls to the `wheel` package and delete the duplicates.

---

## 3. Enable Dependabot for Go modules
**Severity:** high  **Persona(s):** security (ops-sre noted)  **Labels:** dependencies, effort:low

`.github/dependabot.yml` only covers the `github-actions` ecosystem. Go dependency drift is caught only reactively, by govulncheck.

**Suggested approach:** add a `gomod` entry with a weekly schedule to `.github/dependabot.yml`.

---

## 4. Add SECURITY.md with a disclosure policy
**Severity:** high  **Persona(s):** security  **Labels:** documentation, security

The repo is public and has no documented way to report vulnerabilities.

**Suggested approach:** add a short SECURITY.md that points to GitHub private vulnerability reporting and states the supported version (latest release).

---

## 5. Add CONTRIBUTING.md and issue/PR templates
**Severity:** high  **Persona(s):** dx  **Labels:** documentation

The repo has `good first issue` and `help wanted` labels and uses a PR workflow, but has no contributor guide or templates.

**Suggested approach:** write a short CONTRIBUTING.md (build/test/vet commands, conventional commits, PR flow) and add minimal `.github/ISSUE_TEMPLATE/` and `pull_request_template.md` files. Keep them proportionate to the project's size.

---

## 6. Resolve partially tracked prompts/ directory
**Severity:** medium (cross-flagged)  **Persona(s):** security, maintainability, dx  **Labels:** documentation

`.gitignore` excludes `prompts/`, yet 11 of the 23 files under `prompts/completed/` are still tracked from before the ignore rule was added. Nothing explains which state is intended.

**Suggested approach:** decide one way or the other. Either `git rm --cached -r prompts/` to untrack them all, or drop the ignore rule and track them all as design history.

---

## 7. Document dev tooling and debug features
**Severity:** medium (cross-flagged)  **Persona(s):** dx, ops-sre, maintainability  **Labels:** documentation, effort:low

These are hard to discover:
- The Makefile targets are not mentioned in README.
- `cmd/biascheck` is not mentioned in README or CLAUDE.md.
- The `D` debug overlay and `debug_spin_*.log` output (`main.go:296-298, 538-589`) are missing from README's controls table.
- CLAUDE.md says outcomes use crypto/rand, but `main.go:755` also feeds a fixed-seed LCG (`prngState = 42`) into the initial wheel speed.

**Suggested approach:**
- Add a Development section to README (make targets, biascheck).
- Add `D` to the controls table.
- Add biascheck to the CLAUDE.md package list.
- Either switch `randomByte` to crypto/rand (or a randomly seeded source), or document why a fixed seed is acceptable.
