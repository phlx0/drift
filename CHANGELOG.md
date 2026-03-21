# Changelog

All notable changes to drift will be documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
drift uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added

- **rosepine** theme — Rosé Pine colorscheme; palette uses Love, Iris, Foam, and Pine accents with the base background layers as dim variants

---

## [0.5.0] — 2026-03-21

### Added

- **life** scene — Conway's Game of Life on a toroidal grid; newborn cells flash bright, age through the theme palette, then fade to dim variants for visual depth. Auto-resets after `reset_seconds` or after 3 seconds of stagnation. Configurable via `[scene.life]`: `density`, `speed`, `reset_seconds`
- **OLED pixel shift** — the engine nudges the entire rendered image by one cell every 10 seconds, cycling through a 3×3 grid (90-second full cycle). Reduces burn-in risk on OLED displays; all scenes benefit automatically with no per-scene changes. Closes #14

---

## [0.4.1] — 2026-03-21

### Added

- **pipes config**: `[scene.pipes]` now configurable — `heads`, `turn_chance`, `speed`, `reset_seconds`
- **maze config**: `[scene.maze]` now configurable — `pause_seconds`, `fade_seconds`, `speed`

### Fixed

- `drift config --init` template was missing `[scene.pipes]` and `[scene.maze]` sections
- `wildberries` theme was missing from the theme comment in the default config

---

## [0.4.0] — 2026-03-21

### Fixed

- **Scene config**: settings under `[scene.waveform]`, `[scene.rain]`, `[scene.constellation]`, and `[scene.particles]` in `config.toml` were parsed but never applied — scene constructors now receive and use their config values

---

## [0.3.1] — 2026-03-20

### Changed

- **CLI**: `--duration` flag now uses cobra's `Changed()` API instead of a `-1` sentinel to detect explicit user input; flag default changed from `-1` to `0`

---

## [0.3.0] — 2026-03-20

### Added

- **pipes** scene — box-drawing pipes snake across the screen in theme colors, wrapping at edges and resetting every 45 seconds
- **maze** scene — a perfect maze builds itself via depth-first search, rendered as thin box-drawing lines; holds on completion, then dissolves and regenerates

---

## [0.2.0] — 2026-03-20

### Fixed

- **bash shell integration**: terminal no longer freezes when drift activates.
  Background subshells spawned with job control enabled are placed in their own
  process group; bash would then send `SIGTTOU` to the process when drift called
  `tcsetattr`, suspending it and locking the terminal. The timer subshell is now
  spawned with `set +m` so it stays in the shell's process group. A foreground
  guard (`ps` stat check) was also added to skip activation if a command is
  running when the timer fires.

### Added

- `CODE_OF_CONDUCT.md` — Contributor Covenant 2.1
- `SECURITY.md` — private vulnerability reporting via GitHub Security Advisories
- `.github/CODEOWNERS` — all PRs auto-assigned to @phlx0
- `.github/dependabot.yml` — weekly automated updates for GitHub Actions and Go modules
- CI: `windows-latest` added to the test matrix
- CI: `go mod tidy` check to catch uncommitted go.sum drift
- Release: `go test -race` step runs before GoReleaser to gate broken releases

### Changed

- `CONTRIBUTING.md` — expanded with Conventional Commits format, branch naming conventions, and type reference table
- CI: push trigger narrowed to `main` only (PRs remain covered by `pull_request`)
- CI: lint job now derives Go version from `go.mod` instead of a hardcoded value
- CI: `golangci-lint` pinned to `v1.64.0` instead of floating `latest`
- CI: coverage upload targets Go `1.24` (latest in matrix)
- Release: `goreleaser-action` pinned to `v6.3.0`

---

## [0.1.0] — 2026-03-19

First public release.

### Scenes

- **constellation** — stars drift slowly across the screen, connecting nearby neighbours with dotted lines; brightness twinkles per star
- **rain** — columns of half-width katakana characters and digits fall at varying speeds with bright heads and fading trails
- **particles** — a sinusoidal flow field drives 120 glyphs across the screen, leaving ghost trails as they move
- **waveform** — three layered sine waves rendered with Unicode braille characters for sub-character precision; amplitudes breathe in and out independently

### Themes

Seven built-in color themes matched to popular terminal colorschemes: `cosmic`, `nord`, `dracula`, `catppuccin`, `gruvbox`, `forest`, `mono`

### Shell integration

Idle detection via native shell mechanisms — no background daemons:

- **zsh** — TMOUT + TRAPALRM
- **bash** — PROMPT_COMMAND with a background timer
- **fish** — `fish_prompt` / `fish_preexec` event hooks

Activate with `eval "$(drift shell-init zsh)"` (or bash/fish).

### CLI

- `drift --scene <name>` — lock to a specific scene
- `drift --theme <name>` — override the color theme
- `drift --duration <n>` — seconds per scene when cycling, 0 = no cycling
- `drift list scenes` — list available scenes
- `drift list themes` — list themes with live color swatches
- `drift config --init` — write default config to `~/.config/drift/config.toml`
- `drift shell-init zsh|bash|fish` — print shell integration snippet

### Distribution

- Single static binary, no CGO, no runtime dependencies
- Pre-built releases for macOS and Linux (amd64 + arm64)
- goreleaser pipeline with SHA-256 checksums

[0.5.0]: https://github.com/phlx0/drift/releases/tag/v0.5.0
[0.4.1]: https://github.com/phlx0/drift/releases/tag/v0.4.1
[0.4.0]: https://github.com/phlx0/drift/releases/tag/v0.4.0
[0.3.1]: https://github.com/phlx0/drift/releases/tag/v0.3.1
[0.3.0]: https://github.com/phlx0/drift/releases/tag/v0.3.0
[0.2.0]: https://github.com/phlx0/drift/releases/tag/v0.2.0
[0.1.0]: https://github.com/phlx0/drift/releases/tag/v0.1.0
