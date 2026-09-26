# Developer Experience Review — 2026-09-26

**Verdict:** needs-attention

**Project contributor scope (for context):** public-OSS (small/hobby scale) — single active maintainer, but repo carries public-facing signals of inviting outside contributions (GitHub issue label vocabulary includes `good first issue` and `help wanted`, PR-based workflow with numbered merges, dependabot, multi-platform CI).

A newcomer who only reads the README would be productive quickly: it states what the project is, links a screenshot, lists prerequisites (Go 1.27+, a linked Ebitengine install doc for Linux), and gets to a running app in three commands (`go mod tidy`, `go build`, `./roulette-wheel`). That first-five-minutes experience is genuinely good. The friction shows up one layer deeper: the Makefile that actually defines the project's canonical build/run/clean workflow is never mentioned in the README, there is no contribution guide despite labels signaling an interest in outside contributors, a second Go binary (`cmd/biascheck`) exists with no explanation of what it does or why, and there is no test suite or lint command a new contributor could learn the house standard from.

## Findings

**[HIGH] No contribution path despite signals of wanting outside contributors**
- Evidence: No `CONTRIBUTING.md`, no `.github/ISSUE_TEMPLATE/`, no `.github/pull_request_template.md` (confirmed via `find .github -type f`, which returns only `dependabot.yml` and two workflow files). Yet the repository's label vocabulary includes `good first issue` and `help wanted`, and the commit history shows a PR-based merge workflow (`Merge pull request #11...`, `#10...`, `#9...`).
- Why it matters: A first-time contributor who sees a "good first issue" label and opens the repo has no documented path for how to propose a change, what coding conventions to follow, what checks a PR must pass locally before pushing, or what belongs in a PR description. They would have to reverse-engineer expectations from commit history and CI YAML.
- Suggested action: Add a short `CONTRIBUTING.md` covering: how to build/run/vet locally, the commit message convention already in use (Conventional Commits, visible in the log), and what CI checks a PR must pass. A minimal PR template referencing the same checklist would close most of the gap.

**[MEDIUM] Makefile is the documented build interface in CLAUDE.md but invisible in README**
- Evidence: `CLAUDE.md` documents `make build`, `make run`, `make clean` as the canonical commands, but `README.md`'s "Build" section only shows raw `go build -o roulette-wheel .` / `./roulette-wheel`, with no mention of the Makefile at all.
- Why it matters: A human contributor (as opposed to an AI agent reading CLAUDE.md) has no way to discover the Makefile targets from the project's own user-facing documentation. They'd build correctly via the raw `go build` path shown in the README, but would miss the `dist/` output convention and `make clean`, and would have two competing ways of doing the same thing with no cross-reference between them.
- Suggested action: Add a one-line mention in the README's Build section pointing to `make build`/`make run` as the primary path, keeping the raw `go build` line as the "or directly" fallback (which is already how CLAUDE.md frames it).

**[MEDIUM] `cmd/biascheck` is an undocumented internal tool**
- Evidence: `cmd/biascheck/main.go` (94 LOC) runs a Monte-Carlo-style simulation of the ball physics (`runBiasTest`, looping `numSpins` times to tally slot/number frequency) but has no package doc comment, no README section, and no Makefile target. Nothing outside the source file explains it exists, what it's for, or how to invoke it (`go run ./cmd/biascheck`, presumably, but that's a guess).
- Why it matters: A new contributor browsing the tree finds a second `main` package with no onboarding thread connecting it to anything — they must read the source to learn it's a fairness-check utility for the ball's RNG/physics, and would have to guess the invocation and any CLI flags.
- Suggested action: Add a package doc comment stating purpose and usage (e.g., `go run ./cmd/biascheck [numSpins]`), and one sentence in the README or CLAUDE.md architecture section pointing at it — CLAUDE.md's package list already covers `wheel/`, `ball/`, `stats/`, `audio/` but omits `cmd/biascheck` entirely.

**[MEDIUM] No test suite, no test command, no lint/format command documented anywhere**
- Evidence: Snapshot confirms 0 `*_test.go` files across the repo. The Makefile has only `build`, `clean`, `run` — no `test` or `lint` target. CI (`build.yml`) runs `go vet ./...` and `govulncheck`, but never `go test` or a linter (no `golangci-lint` config file found). No `gofmt`/`go fmt` step is documented anywhere.
- Why it matters: There's no established local pattern for a new contributor to imitate when adding tests, and no single command to learn "this is how the project checks a change is good" before pushing — they'd have to infer standards purely from CI YAML, and even then CI itself doesn't test or lint, so there's no example to follow at all. This also means a broken change to `ball/ball.go` physics (deterministic-except-for-RNG state machine) has no regression safety net visible to a newcomer.
- Suggested action: At minimum, document (in CLAUDE.md or CONTRIBUTING) what "done" looks like — even if that's just `go vet ./...` and `gofmt -l .` today — so a contributor knows the bar without guessing. Given the physics/RNG core, even one example unit test would establish the pattern for others to copy.

**[LOW] `prompts/` directory tracked despite being gitignored, with no explanation**
- Evidence: `.gitignore` lists `prompts/`, yet the snapshot notes 11 of 23 on-disk files under `prompts/completed/` are tracked in git anyway (force-added). No README, CLAUDE.md, or CONTRIBUTING mentions what `prompts/` is or why some of its contents are checked in against the ignore rule.
- Why it matters: A contributor inspecting the tree sees a directory that's simultaneously gitignored and partially version-controlled, with no doc explaining the convention (looks like an AI-assisted-development artifact log). This is a small "why does this exist" stumble, not a blocker.
- Suggested action: A one-line note (in CLAUDE.md, since it's dev-tooling metadata rather than user docs) explaining what `prompts/completed/` is for and why some files are force-tracked would remove the head-scratch.

## Notes

- In-code documentation is a genuine strength: `wheel/wheel.go` and `ball/ball.go` both open with a package-level doc comment, and non-obvious constants (e.g., `FrictionCoefficient`, `DeflectorCooldownFrames`) carry inline explanations of their tuning purpose. This is well above what the LOC count alone would suggest.
- The README's prerequisite/quick-start section is right-sized for the project's scope — stating the Go version and linking out to platform-specific Ebitengine setup docs rather than duplicating them is a reasonable choice, not a gap.
- No `.editorconfig` was found, but since this is a single-language Go project, `gofmt` conventions typically substitute for that; not flagged as a separate finding.
- CI actions are pinned to commit SHAs (visible in `build.yml`), which is good supply-chain hygiene, though that's outside this review's scope (see `engineering-ops-sre`/security review for that angle).

### Summary counts
critical=0 high=1 medium=3 low=1
