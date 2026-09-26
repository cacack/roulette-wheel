# Maintainability Review — 2026-09-26

**Verdict:** needs-attention

The codebase is small (3,789 tracked Go LOC across 7 packages), single-maintainer, and otherwise clean — no dead code, no commented-out blocks, no `deprecated`/`legacy` markers, no TODO/FIXME/XXX/HACK tags at all, and consistent naming/comment conventions across `wheel`, `ball`, `stats`, and `audio`. The carrying-cost risk is concentrated in one place: there is no automated verification of any kind (zero `*_test.go` files, no `go test` in CI, no lint step), which is a real gap for a project whose entire premise is physics-driven fairness (38-slot outcome distribution). A handful of smaller drift items (partial-tracking of the `prompts/` directory, an undocumented second randomness mechanism, a stale absolute path in a tracked research doc) add minor friction but are cheap to fix. None of this blocks maintenance today, but the test gap will compound as soon as a second contributor touches `ball/` or `wheel/` physics.

## Findings

**[HIGH] Zero automated tests for a physics-fairness-critical codebase**
- Evidence: `find . -name '*_test.go'` returns nothing (confirmed in snapshot: "Test files (`*_test.go`): 0"). The Makefile defines only `build`, `clean`, `run` — no `test` target. `.github/workflows/build.yml` runs `go vet` and `govulncheck` but has no `go test` step. The one piece of correctness tooling that exists, `cmd/biascheck/main.go` (a 94-line simulator that spins the ball thousands of times and prints per-slot hit counts to check for wheel bias), is never invoked in CI — it's a manual, human-read diagnostic with no assertions, not a test. Meanwhile `ball/ball.go` (606 LOC, 5-phase state machine, crypto/rand-driven physics) and `wheel/wheel.go` (1,355 LOC) carry all of the logic whose correctness the whole app depends on (a "casino night" wheel that's visibly unfair is the worst possible failure mode).
- Why it matters: the team has already invested in bias-checking tooling (commit `18b6390 fix(biascheck): match fixed-timestep ball.Update signature`) showing fairness is a known concern, but that investment isn't wired into anything repeatable — a future change to friction constants, deflector hit-zones, or the phase state machine can silently reintroduce bias or a stuck/never-settling ball with no automated signal, only a human noticing during manual play.
- Suggested action: add a `go test` step to `build.yml` and convert `biascheck`'s distribution check into a real `_test.go` with a chi-squared or simple tolerance-band assertion (e.g., each of the 38 slots within X% of expected over N spins), even if unit tests for rendering code stay out of scope.

**[MEDIUM] `prompts/` directory is inconsistently tracked**
- Evidence: 23 prompt files exist on disk under `prompts/completed/`, but `git ls-files prompts/` returns only 11. `.gitignore` added `prompts/` in commit `d6f1dcc` (Dec 29, 2025), one day after the initial commit `b1805a9` which had already tracked the first 11 prompt files. All 12 prompt files created after that point are correctly untracked per `.gitignore`, but the original 11 remain tracked forever (git doesn't retroactively untrack already-tracked files matching a new ignore rule).
- Why it matters: anyone browsing the repo sees a partial, arbitrary-looking slice of project history (only the earliest prompts) with no note explaining why some prompts are committed and most aren't. A future contributor won't know whether to `git add -f` new prompts or leave them out, and cleanup (untracking the stale 11) is easy to forget indefinitely.
- Suggested action: either fully untrack `prompts/completed/*` (`git rm --cached`) for consistency, or explicitly document in CLAUDE.md/README why the first 11 are kept as historical record and the rest aren't.

**[MEDIUM] Two undocumented randomness mechanisms coexist**
- Evidence: `ball/ball.go` uses `crypto/rand` (via `math/big`) for slot-outcome randomness — this is the mechanism CLAUDE.md documents ("Uses crypto/rand for unpredictable results"). `main.go` separately implements a hand-rolled LCG (`prngState`/`randomByte`, lines 754-760) used to jitter the wheel's initial spin speed (`WheelSpinSpeed * (0.8 + randomByte()/255*0.4)`), which also feeds into where the ball ultimately settles relative to the wheel.
- Why it matters: CLAUDE.md's architecture summary implies all spin unpredictability comes from the CSPRNG in `ball/`, but the wheel's own starting-speed variance — a second input into the final outcome — comes from a trivially predictable LCG seeded with a fixed constant (`42`). This isn't flagged anywhere as an intentional trade-off, so the doc and the code tell two different stories about how "unpredictable" the game actually is, and a future contributor fixing "randomness" bugs may only look in `ball/`.
- Suggested action: either move the wheel-speed jitter onto the same crypto/rand path for consistency, or add a one-line comment/CLAUDE.md note explaining why non-critical variance intentionally uses a cheap PRNG while outcome-determining randomness uses crypto/rand.

**[LOW] Tracked research doc contains a stale developer-machine path**
- Evidence: `research/audio-options.md` embeds the literal path `./audio/audio.go` in its body text.
- Why it matters: this reference is only ever valid on the original author's machine; for any other contributor (or the same author on a different machine/clone location) it's a dead pointer that adds no value and will look confusing years from now with no context for why an absolute path is in a design doc.
- Suggested action: replace with the repo-relative path `audio/audio.go`.

**[LOW] Unflagged "not implemented" stub outside the TODO convention**
- Evidence: `main.go:307`, inside the Escape-key handler: `// Could implement exit confirmation here` with an empty else-branch, present unchanged since the initial commit (`b1805a9`, 2025-12-28, ~9 months old per `git blame`).
- Why it matters: the codebase has zero TODO/FIXME/XXX/HACK tags anywhere, which is otherwise a clean signal — but it means this one deferred-feature stub is invisible to any TODO-based audit or grep. If more of these accumulate without a consistent marker, they'll be undiscoverable except by reading every file.
- Suggested action: either tag it `// TODO: exit confirmation dialog` for discoverability, or remove the comment if the no-op behavior is intentional and final.

## Notes
- No other doc/code divergence found: README's build instructions (`go mod tidy`, `go build -o roulette-wheel .`, Go 1.27+ requirement) match `go.mod` (`go 1.27.1`) and actual toolchain. CLAUDE.md's package/state-machine descriptions match the code structure (`NumberSequence`, `RedNumbers`, ball phases `PhaseIdle`→`PhaseSettled`) with the one randomness nuance noted above.
- No commented-out code blocks, no parallel old/new implementations, and no naming-convention drift were found across `wheel`, `ball`, `stats`, `audio`, and `cmd/biascheck` — for a codebase this size and age, that consistency is a positive signal worth preserving as it grows.
- The two large single-file packages (`wheel/wheel.go` at 1,355 LOC, `main.go` at 773 LOC) are approaching a size where splitting would ease navigation, but this is a file-organization/architecture concern, not flagged as a maintainability finding here.

### Summary counts
critical=0 high=1 medium=2 low=2
