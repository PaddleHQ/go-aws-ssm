# Composite actions

Shared workflow building blocks. Listed here with a one-line purpose and, for actions that own a cache, the contract that lets you avoid layering parallel caches on top.

| Action | Purpose |
|---|---|
| [`setup-go`](./setup-go/action.yml) | Install Go (via `actions/setup-go`), authenticate to private PaddleHQ modules, run `go mod download all`. **Owns the Go deps cache** (see below). |
| [`setup-databases`](./setup-databases/action.yml) | Start Postgres (+ MySQL on opt-in), create the `runner` role and `testdatabase`, export `TESTAMENT_POSTGRES_DSN`. **Restores the Postgres template cache** (see below). |
| [`snapshot-postgres-templates`](./snapshot-postgres-templates/action.yml) | Tar the PG data dir at job end so `actions/cache`'s post-step uploads it. **Saves the Postgres template cache** (the other half of the lifecycle owned by `setup-databases`). |
| [`lint-scope`](./lint-scope/action.yml) | Decide whether `golangci-lint` should run, and if so against which packages. Outputs `skip`, `pkgs`, `new-from-rev`. Used by `go-lint.yml` (main) and `go-lint-experimental.yml` (different `config-glob` input, same logic). |
| [`otel-export`](./otel-export/action.yml) | Export the workflow trace to Honeycomb at job end. |
| [`resolve-generated-paths`](./resolve-generated-paths/action.yml) | Expand the `generated-paths` patterns into a concrete list (used by coverage filtering). |
| [`trigger-automerge`](./trigger-automerge/action.yml) | Kick the automerge workflow once required checks pass. |
| [`automerge-skipped-comment`](./automerge-skipped-comment/action.yml) | Comment on a PR when automerge declined to run. |

## Cache topology — one owner per artefact

Two caches are managed by composite actions in this repo. **Do not add parallel `actions/cache` steps targeting these paths in any workflow** — they collide on tar extract (`Cannot open: File exists`), `actions/cache` marks the restore as failed, and the post-step saves a fresh ~1.2 GB cache on every run for nothing. This bug has been fixed four times already (lint #483, validate #485, test.yml's `Cache test results` (#487), and inside `actions/setup-go` itself via `cache: false` (#487)).

### Go modules + build cache

| | |
|---|---|
| Paths | `~/go/pkg/mod`, `~/.cache/go-build` |
| Key | `setup-go-${{ runner.os }}-${{ runner.arch }}-go-${{ go-version }}[-<cache-suffix>]-${{ hashFiles('**/go.sum') }}` |
| Owner | [`setup-go`](./setup-go/action.yml) — explicit `actions/cache@v5.0.5`; `actions/setup-go`'s built-in cache is disabled (it 409s on every primary-key hit) |
| Used by | every workflow that needs Go — lint, validate, build, fuzz (default key); test (`cache-suffix: cover`); race and test-combined (`cache-suffix: race`) |

`setup-go` takes a `cache-suffix` input. Jobs that compile with non-default flags pass a flavour so each writer owns its own key — race and test-combined pass `race`, test passes `cover`, everything else stays empty. Without per-flavour keys, parallel jobs collide on a single key and only one save wins per run, leaving the others permanently cold.

**Rule:** call `setup-go` (with the right `cache-suffix` if you compile with non-default flags). Don't write a second `actions/cache` step for these paths.

### Postgres template DB

| | |
|---|---|
| Path | `/tmp/pg-template.tar.zst` |
| Key | `${{ runner.os }}-pg${{ pg-major }}-template-testament${{ testament-version }}-${{ migration-hash }}` |
| Restore owner | [`setup-databases`](./setup-databases/action.yml) |
| Save owner | [`snapshot-postgres-templates`](./snapshot-postgres-templates/action.yml), called with `if: always()` at job end |
| Used by | test, test-combined, race |

**Why two composites for one cache?** GitHub composite actions can't declare post-steps. setup-databases restores the tar before PG starts; the snapshot composite tars again at job end so `actions/cache`'s auto-save uploads on cache-miss. Two composites — one for each end of the lifecycle.

**Rule:** if you add a new job that uses testament, call `setup-databases` at the start *and* `snapshot-postgres-templates` at the end. Don't bypass either side.

## Caches not owned here

| Artefact | Owner |
|---|---|
| `~/.cache/golangci-lint` (lint analysis cache) | `golangci/golangci-lint-action`'s built-in cache (in `go-lint.yml`) |
| `~/.cache/go-build/fuzz` (fuzz corpus) | inline `actions/cache` step in `fuzz.yml` (independent sub-path, no collision risk) |

These are noted for completeness — they don't conflict with anything above.
