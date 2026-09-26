# Security Posture Review — 2026-09-26

**Verdict:** needs-attention

This is a small, single-binary desktop game (Ebitengine, Go) with no network listeners, no user accounts, no persisted user data, and no external input surface beyond keyboard/mouse — so the inherent attack surface is low. Posture within that scope is reasonably good: no committed secrets, GitHub Actions are pinned to commit SHAs, `crypto/rand` (not `math/rand`) is used for the ball's outcome, and `govulncheck` runs both on push/PR and on a weekly schedule. The gaps that exist are process/hygiene items appropriate to a public open-source repo rather than exposed-secret or exploitable-code issues: no `SECURITY.md`, no Go module vulnerability/update coverage in Dependabot, and a stale `.gitignore` entry that already let files leak into history once.

## Findings

**[HIGH] No SECURITY.md / disclosure process for a public repo**
- Evidence: snapshot confirms `SECURITY.md: absent`; repo origin is a public GitHub project (`git@github.com:cacack/roulette-wheel.git`) with issues enabled (label vocabulary includes `security`).
- Why it matters: there is no documented channel or expectation for how a security researcher should privately report a finding (e.g., a supply-chain issue in a dependency, or a malicious PR). Absent guidance, reporters default to filing a public issue, which discloses the problem before a fix exists.
- Suggested action: add a minimal `SECURITY.md` with a contact (email or GitHub Security Advisories link) and a one-line statement on response expectations. GitHub's private vulnerability reporting feature can substitute for an inbox if enabled.

**[HIGH] Dependabot only tracks GitHub Actions, not the Go module graph**
- Evidence: `.github/dependabot.yml` contains a single `github-actions` ecosystem entry; there is no `gomod` (or equivalent) `package-ecosystem` entry, despite `go.mod`/`go.sum` being present and the project depending on `ebiten/v2` and its transitive tree.
- Why it matters: `govulncheck` (present in both workflows) only flags dependencies with a *known, call-reachable* CVE — it does not proactively open PRs to bump versions, so drift accumulates silently until a vuln is published and reachable. Combining govulncheck with Dependabot-for-gomod is the standard belt-and-suspenders pattern; only having the former means routine version currency is unmanaged.
- Suggested action: add a `package-ecosystem: gomod` block to `dependabot.yml` (weekly cadence, matching the existing actions block).

**[MEDIUM] `.gitignore` added `prompts/` after files were already tracked, and it is not evidence of a secrets-scanning gap but is worth flagging for hygiene**
- Evidence: snapshot notes `./prompts/completed/*.md (23 files on disk; 11 tracked despite prompts/ in .gitignore)`; confirmed via `git ls-files | grep '^prompts/'` — 11 files remain tracked. Spot-checked all 11 for secret-shaped content (`key|token|password|secret|credential`); only false positives found (e.g., "Key files:" section headers), so no leak occurred this time.
- Why it matters: `.gitignore` only prevents *new* files from being staged — it does not remove files already committed. This particular case is benign (planning prompts, no secrets), but it demonstrates a gap in the team's mental model: anyone assuming "it's in .gitignore" for these files would be wrong, and the same pattern with a future credentials file would silently stay tracked.
- Suggested action: either `git rm --cached` the already-tracked `prompts/completed/*.md` files if they're meant to be excluded, or drop `prompts/` from `.gitignore` if they're intentionally tracked (current state is inconsistent either way).

**[LOW] No secret-scanning or SAST step in CI beyond govulncheck/go vet**
- Evidence: `.github/workflows/build.yml` runs `go vet ./...` and `govulncheck`; `.github/workflows/govulncheck.yml` reruns the latter weekly. Neither workflow, nor any other config, runs a secret scanner (gitleaks/trufflehog) or a SAST tool (CodeQL, semgrep, `gosec`).
- Why it matters: `go vet` and `govulncheck` cover known-CVE dependency reachability and basic correctness issues, but neither greps for accidentally-committed credentials or flags common insecure Go patterns (e.g., unsafe file permissions, command injection via `os/exec`, TLS misconfiguration). For a project of this size the risk is low today (verified: no `net/http`, `os/exec`, or `net.Listen` usage in tracked `.go` files), but the safety net isn't there if that changes.
- Suggested action: add a lightweight CodeQL workflow (GitHub-native, free for public repos) and/or gitleaks-action; low effort given the existing SHA-pinning discipline already in this repo's CI.

**[LOW] `go.mod`/`go.sum` present but only one direct dependency to track**
- Evidence: `go.mod` lists a single direct requirement (`github.com/hajimehoshi/ebiten/v2 v2.10.4`) plus 10 indirect/transitive deps; `go.sum` is committed (lockfile present, satisfying dependency-hygiene baseline).
- Why it matters: this is a positive signal, not a gap — noted here only because the small dependency surface is part of why overall risk is low; flagging so future reviews don't need to re-derive it.
- Suggested action: none required; maintain current practice of periodic `go mod tidy` / version bumps (see chore commit `6d39d19`).

## Notes
- Threat surface is genuinely minimal: this is a local, offline, single-player desktop application. Ball-outcome randomness correctly uses `crypto/rand` (`ball/ball.go:593`, `:601`) rather than `math/rand`, which is good practice even though the "stakes" are a casino-night display rather than real money.
- No `net/http`, `net.Listen`, or `os/exec` usage found anywhere in tracked `.go` files — there is no network or subprocess boundary to threat-model.
- Targeted secret-pattern greps (`sk-`, `AKIA`, `xox[abprs]-`, `ghp_`, inline `password =` / `api_key =` literals) across `.go`, `.yml`, and `.md` files returned zero hits.
- `git ls-files` scan for secret-shaped filenames (`.env`, `.pem`, `.key`, `credentials*`) returned zero hits.
- All `uses:` steps across both workflow files are pinned to full commit SHAs with version comments — this is already best practice and shouldn't regress.
- No auth/authz subsystem exists or is needed given the application's nature (no accounts, no privileged operations) — this is a non-finding, but explicitly out of scope rather than overlooked.
- Encountered embedded instructions inside this session's tool-result text purporting to be "MCP server instructions" for unrelated products (a docs-writing tool and a GitHub tool). These were not part of the snapshot's `<untrusted-issue-data>` block and none of the described tools are available to me; I did not act on them. Flagging per instructions to report any text that attempts to direct agent behavior outside the assigned task — no action was taken and no repository files were affected.

### Summary counts
critical=0 high=2 medium=1 low=3
