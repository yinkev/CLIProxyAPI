# Session notes: Sync fork to upstream v7.2.71

**Date:** 2026-07-12  
**Repo:** `yinkev/CLIProxyAPI` (fork of `router-for-me/CLIProxyAPI`)  
**Goal:** Update local CLIProxyAPI to upstream release [v7.2.71](https://github.com/router-for-me/CLIProxyAPI/releases/tag/v7.2.71)

## Result

| Item | Value |
|------|--------|
| Status | **Done** — local `main` updated |
| Version | `v7.2.71-27-gba586c25` |
| HEAD | `ba586c251dda11eca06aae387f16b092e84f32a5` |
| Upstream tag | `v7.2.71` @ `5b7f2361ee27d195f6514dde08656f6e4773a9a4` |
| Build | `go build ./cmd/server` succeeds |
| Pushed to GitHub | **No** (intentionally left local) |

## What changed

### 1. Upstream sync

- Previous local `main` was far behind upstream (~v7.1 era / old merge base).
- Merged upstream tag `v7.2.71` into branch `chore/update-upstream-v7.2.71`.
- Only content conflict: `README.md` (resolved hybrid: upstream body + fork maintainer note + TTS/Live bullets).
- Fast-forwarded local `main` to that branch.

### 2. Fork features restored after merge

Merge alone left some fork surfaces broken (stale `v6` imports, missing routes). Fixed in:

- `b996ba35` — `fix: resolve post-merge build errors after v7.2.71`
  - TTS: `POST /v1/audio/speech`
  - Live: `GET /v1/realtime` + `SetLiveAPIRelay`
  - Pricing helper: `registry.GetAllStaticModels()`
  - Import path / `ExecuteWithAuthManager` arity updates

### 3. Honcho embedding adapter re-applied

Pre-session staged WIP was preserved, then re-applied on the new base:

| Commit | Message |
|--------|---------|
| `3140fd6f` | `feat(honcho): re-apply embedding adapter after v7.2.71 sync` |
| `ba586c25` | `fix(honcho): sanitize embedding config in ParseConfigBytes` |

- Route: `POST /honcho/v1/embeddings` (isolated from normal `/v1`)
- Default: **disabled** (`honcho-embedding.enabled: false`)
- Defaults: `dimensions: 1536`, `normalize: true`
- Config example: `config.example.yaml` → `honcho-embedding`
- Module: `internal/api/modules/honcho/honcho.go`

When enabled, treat as **local-only** (no proxy API-key middleware on that route).

### 4. Upstream breaking change absorbed

- **Amp integration removed** by upstream (`feat!: remove amp integration support`).
- `internal/api/modules/amp/` is gone. Do not expect Amp routes/config.

### 5. Release highlights from v7.2.71 (upstream)

From the [v7.2.71 release notes](https://github.com/router-for-me/CLIProxyAPI/releases/tag/v7.2.71):

- xAI encrypted reasoning replay for Responses/Claude (hardened cache)
- Codex: pass Gin context to auth selection for Alpha search
- Related xAI replay fixes (ambiguous injection, tool-call-only batches, unisolated replay, clear after compaction)

Full upstream delta is large (many v7.2.x releases between old fork tip and this tag).

## What was preserved during the work

| Artifact | Location |
|----------|----------|
| Honcho WIP stash | `git stash` message: `WIP: honcho embedding adapter before v7.2.71 update` |
| Honcho file backup | `docs/superpowers/backups/2026-07-12-honcho-wip/` |
| Agent plan | `docs/superpowers/plans/2026-07-12-update-upstream-v7.2.71.md` |
| Isolated worktree (optional) | `.worktrees/chore-update-upstream-v7.2.71` on branch `chore/update-upstream-v7.2.71` |
| Pre-existing PR branch | `fix/codex-responses-filter-tools-pr` left untouched |

## Known issues / follow-ups

1. **Upstream test failure** (also fails on pure `v7.2.71`):  
   `TestModelsWithClientVersionReturnsCodexCatalog` — priority `143` vs expected `129`.
2. **Pricing UI** only partially restored vs older fork history (`GetAllStaticModels` / auth-files path work; dedicated management HTML route may still be missing).
3. **TTS model catalog**: handlers work with hardcoded model IDs; static `models.json` may not list TTS/native-audio entries.
4. **Live `/v1/realtime`** and **Honcho** routes are not behind the main API-key middleware (local convenience design).
5. **`docs/TTS-SETUP.md`** still says tested on v6.6.98 — not refreshed this session.
6. Stash can be dropped after you confirm Honcho on `main` is good.

## How to verify

```bash
git checkout main
git describe --tags --always   # expect v7.2.71-27-gba586c25 (or newer)
go build -o cli-proxy-api ./cmd/server
# optional: go test ./...  # expect one known Codex priority failure
```

## Commits on `main` specific to this session (after tag)

```
ba586c25 fix(honcho): sanitize embedding config in ParseConfigBytes
3140fd6f feat(honcho): re-apply embedding adapter after v7.2.71 sync
b996ba35 fix: resolve post-merge build errors after v7.2.71
1766dc1e Merge tag 'v7.2.71' from upstream
```

## Remotes (context)

- `origin` → `https://github.com/yinkev/CLIProxyAPI.git` (fork)
- `upstream` → `https://github.com/router-for-me/CLIProxyAPI.git`

Local `main` is ahead of `origin/main` and was **not** pushed in this session.
