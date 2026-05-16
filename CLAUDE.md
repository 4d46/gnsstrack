# gnsstrack — Claude Code Project Instructions

This file extends the global preferences in `~/.claude/CLAUDE.md`.
Project-specific rules below take precedence where they overlap.

## Go toolchain — run before every commit

```
gofmt -l .              # must produce no output (formatting clean)
go vet ./...            # must produce no output
go test ./...           # all tests must pass
make build-linux-arm64  # ARM64 binary must compile cleanly
```

These are required, not optional. If any step fails, fix it before committing.

## Branch protection

Main branch is protected — direct pushes are blocked both locally (pre-push hook)
and on GitHub (ruleset). All changes must go through a pull request.

The local hook lives in `.githooks/pre-push`. New checkouts must run:
```
git config core.hooksPath .githooks
```

## Release process

1. Feature/fix work on a descriptive branch → PR → merge to main
2. For a release, create a `release/vX.Y.Z` branch
3. Update `CHANGELOG.md`:
   - Move items from `[Unreleased]` into a new `[vX.Y.Z] - YYYY-MM-DD` section
   - Leave a fresh empty `[Unreleased]` section at the top
4. Commit CHANGELOG, open PR, merge
5. Pull main, then tag:
   ```
   git tag -a vX.Y.Z -m "Release vX.Y.Z"
   git push origin vX.Y.Z
   ```
6. GitHub Actions builds the ARM64 package and publishes the release automatically

## CHANGELOG discipline

Every user-facing change must have a CHANGELOG entry before it is released:
- `### Fixed` — bug fixes
- `### Changed` — behaviour or config changes
- `### Added` — new features
- `### Removed` — removed functionality (include migration notes)

Use semantic versioning: patch for fixes/config, minor for new features,
major for breaking changes.

## GitHub Actions security

- All actions pinned to full commit SHAs with a `# vX.Y.Z` comment
- Dependabot configured for weekly Action updates — review PRs before merging
- `permissions: read-all` at workflow level; `contents: write` on release job only
- `if: github.repository == '4d46/gnsstrack'` guard prevents fork releases
- Tag ruleset restricts `v*` pushes to Repository Admin; tags must be signed

## Target hardware

Single ARM64 binary targets both Raspberry Pi CM4 and CM5 — both are ARMv8.
Cross-compiled with: `GOOS=linux GOARCH=arm64`

Release package contains: binary, `config.yaml`, `gnsstrack.service`, `README.md`
Stable Ansible download URL:
`https://github.com/4d46/gnsstrack/releases/latest/download/gnsstrack-linux-arm64.tar.gz`
