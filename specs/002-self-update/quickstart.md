# Quickstart: Self-Update

**Feature**: 002-self-update | **Date**: 2026-05-30

Five-minute path through every user-visible piece of this feature.

## Prerequisites

- Local: Go 1.26.3+, GoReleaser, `git-chglog` (for release; not needed
  for normal use), `make`.
- For an end-to-end release: write access to `buildoutinc/injector`
  and `GITHUB_TOKEN` configured locally or in CI.

## A. Try the upgrade flow locally (no real release needed)

```bash
make build
./bin/inject version              # prints "inject dev (commit …)"
./bin/inject upgrade --check      # exits 0; says no upgrade available
./bin/inject upgrade              # local "dev" build → "already up to date"
INJECT_NO_UPDATE_CHECK=1 ./bin/inject version   # no notice line
```

Expected:
- `version` shows `dev` because no ldflag was injected.
- `upgrade --check` is silent on the network (cache exists / dev-mode).

## B. Cut a real release

```bash
git checkout main
git pull
# Make some commits with Conventional Commit prefixes:
git commit -m "feat: add inject upgrade"
git commit -m "fix: handle missing cache file"
git commit -m "chore: bump deps"
# Tag and push:
git tag -a v0.3.0 -m "v0.3.0"
git push origin v0.3.0
```

Watch the **Release** workflow in the GitHub Actions tab. When green,
visit Releases and confirm:

- A release named `v0.3.0` exists.
- Body groups commits under `## Features`, `## Fixes`, `## Other Changes`.
  The `chore:` commit shows under "Other Changes".
- Four `.tar.gz` archives + `checksums.txt` attached.

## C. Self-upgrade against the new release

On a machine running an older `inject`:

```bash
inject version            # e.g. inject v0.2.0 (…)
# See the inline notice: "A newer version is available: v0.3.0"
inject upgrade --check    # confirms upgrade is available
inject upgrade            # downloads, verifies, swaps, prints "upgraded v0.2.0 → v0.3.0"
inject version            # now reports v0.3.0
```

Expected: each step prints a single, clear line and exits 0.

## D. Opt out

```bash
export INJECT_NO_UPDATE_CHECK=1
inject version            # no notice line; no network call
inject                    # help screen, no notice
```

## E. Failure modes to verify

- **Offline upgrade**: `airport -z; inject upgrade` → exits non-zero
  with one network sentence on stderr; binary unchanged.
- **Permission denied**: `sudo cp ./bin/inject /usr/local/bin/inject;
  /usr/local/bin/inject upgrade` (as your user) → refuses before
  downloading; tells you to use your package manager.
- **Already current**: re-run `inject upgrade` immediately → "already
  up to date".

## F. Cache inspection (debugging)

```bash
cat "$(go env GOPATH | head -c0)$HOME/.cache/inject/update-check.json"
# {"checked_at":"2026-05-30T14:22:00Z","latest_version":"v0.3.0","binary_version":"v0.3.0"}
rm "$HOME/.cache/inject/update-check.json"
inject version            # re-populates the cache
```
