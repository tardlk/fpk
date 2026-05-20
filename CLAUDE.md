# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Summary

精选 fnOS 第三方应用打包仓库，包含 7 个应用（Plex, Emby, Jellyfin, qBittorrent, Firefox, MoviePilot, OpenList），输出 `.fpk` 安装包。 100% Bash — downloads upstream binaries, merges with shared lifecycle framework, outputs `.fpk` tarballs. Daily CI auto-syncs upstream versions. See `AGENTS.md` for detailed knowledge base.

## Build Commands

```bash
# Build an app locally (from repo root)
cd apps/plex && ./update_plex.sh                        # latest version, auto-detect arch
cd apps/plex && ./update_plex.sh --arch arm              # force ARM build
cd apps/plex && ./update_plex.sh --arch x86 1.42.2.10156 # specific version + arch

# Same pattern for all binary-extraction apps
cd apps/emby && ./update_emby.sh
cd apps/jellyfin && ./update_jellyfin.sh
cd apps/qbittorrent && ./update_qbittorrent.sh
cd apps/openlist && ./update_openlist.sh

# Docker apps (firefox, moviepilot) — no local update script, CI-only builds
# cd apps/firefox

# Generic fpk packager (used by CI, rarely needed locally)
./scripts/build-fpk.sh apps/plex app.tgz [version] [platform]

# Scaffold a new app
./scripts/new-app.sh <name> "<display name>" <port>
```

## Validation

There is no unit test framework. Validation is structural:
- `bash -n <script>` for syntax checking
- `scripts/build-fpk.sh` validates manifest keys, icon files, and directory structure before packaging
- `scripts/test/` contains CI validation scripts: `static-check.sh` (shellcheck/lint), `verify-fpk.sh` (package integrity), `health-schema.sh` (health check endpoints), `run-fpk-tests.sh` / `run-batch.sh` (test runner)
- Functional testing requires a real fnOS device

## Architecture

### Two App Categories

| Type | Apps | build.sh Pattern | Local update_*.sh |
|------|------|-----------------|-------------------|
| Binary extraction | plex, emby, jellyfin, qbittorrent, openlist | Download upstream → extract → repack into app.tgz | Yes |
| Docker | firefox, moviepilot | Copy docker-compose.yaml + ui/ → tar into app.tgz (no download) | No (CI only) |

Docker apps skip the `app.tgz` minimum-size check and use `apps/<app>/fnos/docker/` for their compose file.

### Plugin Contract System

Each app implements a standardized contract in `scripts/apps/<app>/`:

| File | Purpose | Key Output |
|------|---------|------------|
| `meta.env` | Static metadata (sourced) | `FILE_PREFIX`, `RELEASE_TITLE`, `DEFAULT_PORT` |
| `get-latest-version.sh` | Query upstream for latest version | `VERSION=x.y.z` to stdout + `$GITHUB_OUTPUT` |
| `build.sh` | Download upstream + create `app.tgz` | `app.tgz` in cwd |
| `release-notes.tpl` | Release notes template | `${VARIABLE}` placeholders for `envsubst` |

Full contract spec: `scripts/apps/README.md`

### Overlay Packaging Pattern

`.fpk` files are built by layering:
1. `shared/cmd/` — common lifecycle hooks (install/upgrade/uninstall/start/stop)
2. `apps/<app>/fnos/cmd/` — app-specific overrides (wins on conflict)
3. `apps/<app>/fnos/config/`, `ui/`, `wizard/`, icons, manifest
4. `apps/<app>/var/` — optional seed files synced to `TRIM_PKGVAR` by `sync_var_folder` on install
5. `apps/<app>/fnos/*.sc` — port forwarding config

Core lifecycle framework: `shared/cmd/common` (~375 lines). Wizard default: `shared/wizard/`, overridden by app-specific `fnos/wizard/`.

### Local Build Library

`scripts/lib/update-common.sh` provides shared functions (`main_flow`, `detect_arch`, `parse_args`, `build_fpk`, `update_manifest`, etc.) for all `update_*.sh` scripts. Each app's update script defines callbacks (`app_set_arch_vars`, `app_get_latest_version`, `app_download`, `app_build_app_tgz`) then sources the library and calls `main_flow "$@"`.

### CI Pipeline

```
Single entry workflow (build-apps.yml)
  → detect: auto-discover changed apps (push) or all apps (schedule/dispatch)
  → dynamic matrix per app → Reusable workflow (reusable-build-app.yml)
    → check-version: source meta.env + get-latest-version.sh + resolve-release-tag.sh
    → build: matrix [x86, arm] → build.sh + build-fpk.sh
    → release: create GitHub release with release-notes.tpl
```

CI is idempotent — skips if release tag already exists. Version tags: `<app>/v<version>` with `-r2`, `-r3` auto-increment for rebuilds. `scripts/ci/resolve-release-tag.sh` handles tag resolution; `scripts/ci/generate-apps-json.sh` generates the app catalog.

## Key Conventions

- **Language**: Chinese for user-facing messages (info/warn/error), English for code comments
- **Architecture**: Always dual-build x86 + arm. Detect via `uname -m` or `--arch` flag
- **Manifest format**: Fixed-width alignment at column 16 (`appname         = value`). Optional `fpk_version` key via `$FPK_VERSION` env var
- **Color output helpers**: `info()` green, `warn()` yellow, `error()` red + exit
- **Cleanup**: All scripts use `trap cleanup EXIT` for temp dir removal
- **fnOS runtime env vars**: `TRIM_APPNAME`, `TRIM_APPDEST`, `TRIM_PKGVAR`, `TRIM_PKGETC`, `TRIM_PKGHOME`
- **Changelog format**: `apps/<app>/CHANGELOG.md` with `## YYYY-MM-DD` date headers, latest entry first. CI auto-extracts the first section into release notes as `${CHANGELOG}`.

## Anti-Patterns

- Never modify upstream binaries — download and repackage only
- Never hardcode architecture — always use detection or `--arch` flag
- Never duplicate `shared/cmd/` logic in app-specific overrides — only override what differs
- Never skip checksum — `app.tgz` md5 must be written to manifest
- Never create per-app CI build scripts in `scripts/ci/` — use `scripts/apps/<app>/build.sh`
