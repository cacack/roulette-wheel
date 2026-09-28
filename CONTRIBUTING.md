# Contributing

Thanks for your interest! This is a small hobby project, so the process is light.

## Setup

Requires the Go version in `go.mod`. Linux also needs the [Ebitengine dependencies](https://ebitengine.org/en/documents/install.html).

```bash
make run        # Run the app
make test       # Run all tests, including the fairness simulation
make test-short # Skip the fairness simulation
go vet ./...
```

If you change ball or wheel physics, also run `go run ./cmd/biascheck` to confirm results stay fair.

## Workflow

1. Open an issue first for anything beyond a small fix.
2. Branch from `main` and keep each PR focused on one change.
3. Write commits as [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `docs:`, `chore:`, ...).
4. Open a PR against `main`, and link the issue with `Closes #N`. CI must pass before merge.

## Security

Please don't open public issues for vulnerabilities. See [SECURITY.md](SECURITY.md).
